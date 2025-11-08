package fetcher

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestVenueAddressFetcher_BasicFetch(t *testing.T) {
	// Create test server that returns HTML with venue address
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
			<html>
				<body>
					<div class="container">
						<div><div><h2><small></small><small>123 Test Street, Test City</small></h2></div></div>
					</div>
				</body>
			</html>
		`))
	}))
	defer server.Close()

	fetcher := NewVenueAddressFetcher(nil)
	
	address, err := fetcher.Fetch(server.URL, "remote")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if address != "123 Test Street, Test City" {
		t.Errorf("Expected '123 Test Street, Test City', got '%s'", address)
	}
}

func TestVenueAddressFetcher_Caching(t *testing.T) {
	requestCount := int32(0)
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
			<html><body>
				<div class="month"></div>
				<div><div><div>456 Cached Avenue</div></div></div>
			</body></html>
		`))
	}))
	defer server.Close()

	fetcher := NewVenueAddressFetcher(nil)
	
	// First request - should hit server
	address1, err1 := fetcher.Fetch(server.URL, "local")
	if err1 != nil {
		t.Fatalf("First fetch failed: %v", err1)
	}
	
	// Second request - should use cache
	address2, err2 := fetcher.Fetch(server.URL, "local")
	if err2 != nil {
		t.Fatalf("Second fetch failed: %v", err2)
	}
	
	if address1 != address2 {
		t.Errorf("Expected same address, got '%s' and '%s'", address1, address2)
	}
	
	if atomic.LoadInt32(&requestCount) != 1 {
		t.Errorf("Expected 1 HTTP request, got %d", requestCount)
	}
}

func TestVenueAddressFetcher_ConcurrentRequestDeduplication(t *testing.T) {
	requestCount := int32(0)
	
	// Server with delay to simulate slow response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		time.Sleep(100 * time.Millisecond) // Simulate slow response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
			<html><body>
				<div class="container">
					<div><div><h2><small></small><small>789 Parallel Road</small></h2></div></div>
				</div>
			</body></html>
		`))
	}))
	defer server.Close()

	fetcher := NewVenueAddressFetcher(nil)
	
	const goroutines = 10
	var wg sync.WaitGroup
	results := make([]string, goroutines)
	errors := make([]error, goroutines)
	
	// Launch multiple concurrent requests to the same URL
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			address, err := fetcher.Fetch(server.URL, "remote")
			results[idx] = address
			errors[idx] = err
		}(i)
	}
	
	wg.Wait()
	
	// Verify only one HTTP request was made
	if atomic.LoadInt32(&requestCount) != 1 {
		t.Errorf("Expected 1 HTTP request, got %d", requestCount)
	}
	
	// Verify all goroutines got the same result
	for i := 0; i < goroutines; i++ {
		if errors[i] != nil {
			t.Errorf("Goroutine %d got error: %v", i, errors[i])
		}
		if results[i] != "789 Parallel Road" {
			t.Errorf("Goroutine %d got wrong address: '%s'", i, results[i])
		}
	}
}

func TestVenueAddressFetcher_ErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	fetcher := NewVenueAddressFetcher(nil)
	
	_, err := fetcher.Fetch(server.URL, "remote")
	if err == nil {
		t.Fatal("Expected error for 404 response, got nil")
	}
	
	// Verify error is cached
	_, err2 := fetcher.Fetch(server.URL, "remote")
	if err2 == nil {
		t.Fatal("Expected cached error, got nil")
	}
}

func TestVenueAddressFetcher_FetchMultiple(t *testing.T) {
	requestCount := int32(0)
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`
			<html><body>
				<div class="container">
					<div><div><h2><small></small><small>Address %d</small></h2></div></div>
				</div>
			</body></html>
		`, count)))
	}))
	defer server.Close()

	fetcher := NewVenueAddressFetcher(nil)
	
	requests := []VenueRequest{
		{URL: server.URL + "/1", Class: "remote"},
		{URL: server.URL + "/2", Class: "remote"},
		{URL: server.URL + "/3", Class: "remote"},
	}
	
	results := fetcher.FetchMultiple(requests)
	
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
	
	for _, req := range requests {
		if _, ok := results[req.URL]; !ok {
			t.Errorf("Missing result for URL: %s", req.URL)
		}
	}
}

func TestVenueAddressFetcher_ClearCache(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
			<html><body>
				<div class="month"></div>
				<div><div><div>Clear Test Address</div></div></div>
			</body></html>
		`))
	}))
	defer server.Close()

	fetcher := NewVenueAddressFetcher(nil)
	
	// Fetch to populate cache
	fetcher.Fetch(server.URL, "local")
	
	if fetcher.CacheSize() != 1 {
		t.Errorf("Expected cache size 1, got %d", fetcher.CacheSize())
	}
	
	// Clear cache
	fetcher.ClearCache()
	
	if fetcher.CacheSize() != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", fetcher.CacheSize())
	}
}

func TestVenueAddressFetcher_MultipleURLsConcurrent(t *testing.T) {
	requestCounts := make(map[string]*int32)
	mu := sync.Mutex{}
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		if _, ok := requestCounts[r.URL.Path]; !ok {
			var count int32
			requestCounts[r.URL.Path] = &count
		}
		count := requestCounts[r.URL.Path]
		mu.Unlock()
		
		atomic.AddInt32(count, 1)
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`
			<html><body>
				<div class="container">
					<div><div><h2><small></small><small>Address for %s</small></h2></div></div>
				</div>
			</body></html>
		`, r.URL.Path)))
	}))
	defer server.Close()

	fetcher := NewVenueAddressFetcher(nil)
	
	urls := []string{
		server.URL + "/url1",
		server.URL + "/url2",
		server.URL + "/url3",
	}
	
	const requestsPerURL = 5
	var wg sync.WaitGroup
	
	// Make multiple concurrent requests to multiple URLs
	for _, url := range urls {
		for i := 0; i < requestsPerURL; i++ {
			wg.Add(1)
			go func(u string) {
				defer wg.Done()
				_, _ = fetcher.Fetch(u, "remote")
			}(url)
		}
	}
	
	wg.Wait()
	
	// Verify each URL was only fetched once despite multiple concurrent requests
	for path, count := range requestCounts {
		if atomic.LoadInt32(count) != 1 {
			t.Errorf("URL %s was fetched %d times, expected 1", path, atomic.LoadInt32(count))
		}
	}
}

func BenchmarkVenueAddressFetcher_ConcurrentSameURL(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
			<html><body>
				<div class="month"></div>
				<div><div><div>Benchmark Address</div></div></div>
			</body></html>
		`))
	}))
	defer server.Close()

	fetcher := NewVenueAddressFetcher(nil)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = fetcher.Fetch(server.URL, "local")
		}
	})
}
