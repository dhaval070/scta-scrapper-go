package fetcher

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/antchfx/htmlquery"
)

// VenueAddressFetcher fetches venue addresses with caching and request deduplication.
// It ensures only one HTTP request is made per URL, even with concurrent calls.
type VenueAddressFetcher struct {
	client *http.Client
	cache  map[string]*cacheEntry
	mu     sync.RWMutex
}

// cacheEntry holds cached response and in-progress request information
type cacheEntry struct {
	body     string
	err      error
	done     chan struct{} // closed when request completes
	inFlight bool          // true if request is currently in progress
}

// NewVenueAddressFetcher creates a new VenueAddressFetcher with the given HTTP client.
// If client is nil, http.DefaultClient is used.
func NewVenueAddressFetcher(client *http.Client) *VenueAddressFetcher {
	if client == nil {
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	return &VenueAddressFetcher{
		client: client,
		cache:  make(map[string]*cacheEntry),
	}
}

// Fetch retrieves and parses the venue address for the given URL.
// If a request to the same URL is already in progress, it waits for that request to complete.
// Responses are cached to avoid duplicate requests.
// The class parameter determines which XPath selector to use ("remote" or "local").
func (f *VenueAddressFetcher) Fetch(url string, class string) (string, error) {
	cacheKey := url + "|" + class
	return f.fetch(cacheKey, url, class)
}

// fetch is the internal implementation that uses cacheKey for deduplication
func (f *VenueAddressFetcher) fetch(cacheKey, url, class string) (string, error) {
	// Check if we have a cached response or in-flight request
	f.mu.RLock()
	entry, exists := f.cache[cacheKey]
	if exists {
		if !entry.inFlight {
			// Cache hit - return cached response
			f.mu.RUnlock()
			return entry.body, entry.err
		}
		// Request in flight - wait for it
		doneChan := entry.done
		f.mu.RUnlock()

		<-doneChan // Wait for in-flight request to complete

		// Now get the completed result
		f.mu.RLock()
		entry = f.cache[cacheKey]
		f.mu.RUnlock()
		return entry.body, entry.err
	}
	f.mu.RUnlock()

	// No cache entry exists - create one and start the request
	f.mu.Lock()

	// Double-check: another goroutine might have created the entry
	entry, exists = f.cache[cacheKey]
	if exists {
		if !entry.inFlight {
			// Cache hit - return cached response
			f.mu.Unlock()
			return entry.body, entry.err
		}
		// Request in flight - wait for it
		doneChan := entry.done
		f.mu.Unlock()

		<-doneChan

		f.mu.RLock()
		entry = f.cache[cacheKey]
		f.mu.RUnlock()
		return entry.body, entry.err
	}

	// Create new cache entry and mark as in-flight
	entry = &cacheEntry{
		done:     make(chan struct{}),
		inFlight: true,
	}
	f.cache[cacheKey] = entry
	f.mu.Unlock()

	// Perform the HTTP request and scrape address
	body, err := f.scrapeVenueAddress(url, class)
	if err != nil {
		log.Printf("fetcher error: url=%s, err= %v\n", url, err)
	}

	// Update cache entry with result
	f.mu.Lock()
	entry.body = body
	entry.err = err
	entry.inFlight = false
	close(entry.done) // Signal all waiting goroutines
	f.mu.Unlock()

	return body, err
}

// scrapeVenueAddress performs the HTTP request and scrapes the venue address
// This is copied from pkg/parser/parser.go:GetVenueAddress
func (f *VenueAddressFetcher) scrapeVenueAddress(url string, class string) (string, error) {
	var err error
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	var resp *http.Response
	var try int

	for try = 1; try < 3; try += 1 {
		resp, err = f.client.Do(req)
		if err != nil || resp.StatusCode != 200 {
			log.Printf("error in fetcher(retrying): url=%s,  err=%v\n", url, err)
			err = fmt.Errorf("HTTP request failed:  %w", err)
			var backoff = 2 * int64(try) * int64(time.Second)
			time.Sleep(time.Duration(backoff))
			continue
		}
		break
	}

	if err != nil {
		return "", fmt.Errorf("all retry failed %w", err)
	}

	defer resp.Body.Close()

	if try > 1 {
		log.Println("retry successful for url ", url)
	}

	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	var address string

	// theonedb.com - remote URLs
	switch class {
	case "remote":
		item := htmlquery.FindOne(doc, `//div[@class="container"]/div/div/h2/small[2]`)
		if item == nil {
			return "", fmt.Errorf("address node not found for remote URL: %v", url)
		}
		address = htmlquery.InnerText(item)
	case "local":
		// local URLs
		node := htmlquery.FindOne(doc, `//div[@class="month"]/following-sibling::div/div/div`)
		if node == nil {
			return "", fmt.Errorf("address node not found for local URL: %v", url)
		}
		address = htmlquery.InnerText(node)
	default:
		return "", fmt.Errorf("unknown class type: %s (expected 'remote' or 'local')", class)
	}

	return address, nil
}

// VenueRequest represents a venue URL fetch request
type VenueRequest struct {
	URL   string
	Class string
}

// FetchMultiple fetches multiple venue addresses in parallel and returns a map of URL to address.
// All requests are deduplicated and cached.
func (f *VenueAddressFetcher) FetchMultiple(requests []VenueRequest) map[string]string {
	results := make(map[string]string)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, req := range requests {
		wg.Add(1)
		go func(r VenueRequest) {
			defer wg.Done()
			address, err := f.Fetch(r.URL, r.Class)
			if err == nil {
				mu.Lock()
				results[r.URL] = address
				mu.Unlock()
			}
		}(req)
	}

	wg.Wait()
	return results
}

// ClearCache removes all cached entries
func (f *VenueAddressFetcher) ClearCache() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.cache = make(map[string]*cacheEntry)
}

// CacheSize returns the number of cached entries
func (f *VenueAddressFetcher) CacheSize() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.cache)
}
