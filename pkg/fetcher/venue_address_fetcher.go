package fetcher

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/antchfx/htmlquery"
)

// VenueAddressFetcher fetches venue addresses with caching and request deduplication.
// It ensures only one HTTP request is made per URL, even with concurrent calls.
type VenueAddressFetcher struct {
	client       *http.Client
	cache        map[string]*cacheEntry
	mu           sync.RWMutex
	maxCacheSize int
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
		client:       client,
		cache:        make(map[string]*cacheEntry),
		maxCacheSize: 50000,
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
	f.evictIfNeeded()
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
// Rate limiting is handled by the HTTP client's transport layer.
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

	htmlContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read HTML: %w", err)
	}

	doc, err := htmlquery.Parse(bytes.NewReader(htmlContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	var address string

	if strings.Contains(url, "rinkdb.com") {
		nodes := htmlquery.Find(doc, `//div[@class="container-fluid"]//h3`)
		if len(nodes) == 0 {
			log.Printf("address nodes not found %s", url)
			return address, nil
		}
		for _, node := range nodes {
			address = address + htmlquery.InnerText(node) + " "
		}
		return strings.Trim(address, " \n"), nil
	}

	// theonedb.com - remote URLs
	switch class {
	case "remote":
		item := htmlquery.FindOne(doc, `//div[@class="container"]/div/div/h2/small[2]`)
		if item == nil {
			f.logAddressNotFound(url, class, string(htmlContent), "address node not found for remote URL")
			return "", fmt.Errorf("address node not found for remote URL: %v", url)
		}
		address = htmlquery.InnerText(item)
	case "local":
		// local URLs
		node := htmlquery.FindOne(doc, `//div[@class="month"]/following-sibling::div/div/div`)
		if node == nil {
			f.logAddressNotFound(url, class, string(htmlContent), "address node not found for local URL")
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

// evictIfNeeded removes excess cache entries when cache size exceeds maxCacheSize.
// Must be called with f.mu.Lock() held.
func (f *VenueAddressFetcher) evictIfNeeded() {
	if len(f.cache) < f.maxCacheSize {
		return
	}
	// Evict 10% of entries, but keep in-flight requests
	target := f.maxCacheSize * 9 / 10 // reduce to 90% of max
	toDelete := len(f.cache) - target
	if toDelete <= 0 {
		return
	}
	// Iterate over map and delete entries that are not in-flight
	deleted := 0
	for key, entry := range f.cache {
		if !entry.inFlight {
			delete(f.cache, key)
			deleted++
			if deleted >= toDelete {
				break
			}
		}
	}
}

// CacheSize returns the number of cached entries
func (f *VenueAddressFetcher) CacheSize() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.cache)
}

// logAddressNotFound writes the URL and page HTML to a separate log file when address node is not found
func (f *VenueAddressFetcher) logAddressNotFound(url string, class string, htmlContent string, errorMsg string) {
	logFile := "address_not_found.log"
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("failed to open address not found log file: %v", err)
		return
	}
	defer file.Close()

	fmt.Fprintf(file, "###########\n")
	fmt.Fprintf(file, "URL: %s\n", url)
	fmt.Fprintf(file, "Class: %s\n", class)
	fmt.Fprintf(file, "Error: %s\n", errorMsg)
	fmt.Fprintf(file, "HTML Content:\n%s\n", htmlContent)
	fmt.Fprintf(file, "###########\n\n")
}
