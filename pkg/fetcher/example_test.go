package fetcher_test

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"calendar-scrapper/pkg/fetcher"
)

func ExampleVenueAddressFetcher_Fetch() {
	// Create a custom HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create the fetcher
	f := fetcher.NewVenueAddressFetcher(client)

	// Fetch a venue address - first request hits the server
	address1, err := f.Fetch("https://example.com/venue/1", "remote")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("First fetch:", address1)

	// Second fetch - uses cached response, no HTTP request
	address2, err := f.Fetch("https://example.com/venue/1", "remote")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Second fetch (cached):", address2)
}

func ExampleVenueAddressFetcher_FetchMultiple() {
	f := fetcher.NewVenueAddressFetcher(nil)

	// Fetch multiple venue addresses in parallel
	requests := []fetcher.VenueRequest{
		{URL: "https://example.com/venue/1", Class: "remote"},
		{URL: "https://example.com/venue/2", Class: "local"},
		{URL: "https://example.com/venue/3", Class: "remote"},
	}

	results := f.FetchMultiple(requests)

	for url, address := range results {
		fmt.Printf("URL %s: %s\n", url, address)
	}
}

func ExampleVenueAddressFetcher_concurrentDeduplication() {
	f := fetcher.NewVenueAddressFetcher(nil)

	// Multiple goroutines requesting the same URL
	// Only ONE HTTP request will be made, others will wait and share the result
	done := make(chan string, 5)

	for i := 0; i < 5; i++ {
		go func(id int) {
			address, err := f.Fetch("https://example.com/venue/popular", "remote")
			if err != nil {
				done <- fmt.Sprintf("Worker %d: error", id)
				return
			}
			done <- fmt.Sprintf("Worker %d: got address: %s", id, address)
		}(i)
	}

	for i := 0; i < 5; i++ {
		fmt.Println(<-done)
	}
}
