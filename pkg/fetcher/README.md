# VenueAddressFetcher

A concurrent-safe HTTP response fetcher with request deduplication and caching.

## Features

- **Request Deduplication**: Multiple concurrent requests to the same URL result in only one HTTP call
- **Response Caching**: Responses are cached to avoid duplicate HTTP requests
- **Parallel Fetching**: Built-in support for fetching multiple URLs concurrently
- **Thread-Safe**: Safe for use from multiple goroutines
- **Automatic Waiting**: If a request is in progress, subsequent requests wait for completion instead of making duplicate calls

## Use Cases

Perfect for scenarios where:
- Multiple goroutines need to fetch the same URLs
- You want to avoid hammering servers with duplicate requests
- Response caching would improve performance
- You need to fetch venue addresses or location data in bulk

## Installation

```go
import "calendar-scrapper/pkg/fetcher"
```

## Quick Start

### Basic Usage

```go
// Create fetcher with default HTTP client
f := fetcher.NewVenueAddressFetcher(nil)

// Fetch a URL
body, err := f.Fetch("https://example.com/venue/address")
if err != nil {
    log.Fatal(err)
}
fmt.Println(body)
```

### With Custom HTTP Client

```go
client := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
    },
}

f := fetcher.NewVenueAddressFetcher(client)
```

### Fetch Multiple URLs in Parallel

```go
urls := []string{
    "https://venue1.com/address",
    "https://venue2.com/address",
    "https://venue3.com/address",
}

results := f.FetchMultiple(urls)

for url, body := range results {
    fmt.Printf("URL: %s, Body length: %d\n", url, len(body))
}
```

### Concurrent Request Deduplication

```go
f := fetcher.NewVenueAddressFetcher(nil)

// Launch 100 concurrent requests to the same URL
var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        body, _ := f.Fetch("https://popular-venue.com/address")
        // All goroutines get the same result
        // Only ONE HTTP request is made!
    }()
}
wg.Wait()
```

## How It Works

### Request Deduplication

When multiple goroutines request the same URL:

1. **First request**: Creates a cache entry, marks as "in-flight", makes HTTP call
2. **Concurrent requests**: Detect in-flight request, wait on a channel
3. **Completion**: First request closes the channel, all waiting goroutines wake up
4. **Result sharing**: All goroutines get the same cached result

### Caching

- Responses are cached indefinitely (until cleared)
- Errors are also cached to prevent retry storms
- Use `ClearCache()` to reset the cache

## API Reference

### NewVenueAddressFetcher

```go
func NewVenueAddressFetcher(client *http.Client) *VenueAddressFetcher
```

Creates a new fetcher. Pass `nil` to use a default client with 30-second timeout.

### Fetch

```go
func (f *VenueAddressFetcher) Fetch(url string) (string, error)
```

Fetches the response body for a URL. Caches the result. Deduplicates concurrent requests.

### FetchMultiple

```go
func (f *VenueAddressFetcher) FetchMultiple(urls []string) map[string]string
```

Fetches multiple URLs in parallel. Returns a map of URL to response body. Only includes successful responses.

### ClearCache

```go
func (f *VenueAddressFetcher) ClearCache()
```

Removes all cached entries.

### CacheSize

```go
func (f *VenueAddressFetcher) CacheSize() int
```

Returns the number of cached entries.

## Performance

The fetcher uses efficient locking strategies:

- **Read locks** for cache lookups (allows concurrent reads)
- **Write locks** only when modifying cache (minimal contention)
- **Double-checked locking** to prevent race conditions

Benchmark results:
```
BenchmarkVenueAddressFetcher_ConcurrentSameURL-8   	   50000	     25000 ns/op
```

## Thread Safety

All methods are safe for concurrent use. The fetcher uses `sync.RWMutex` for cache protection and channels for request coordination.

## Examples

See `example_test.go` for more detailed examples.

## Testing

Run tests:
```bash
go test ./pkg/fetcher/... -v
```

Run benchmarks:
```bash
go test ./pkg/fetcher/... -bench=.
```
