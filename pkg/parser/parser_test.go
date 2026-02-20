package parser

import (
	"log"
	"testing"

	"github.com/antchfx/htmlquery"
	"github.com/stretchr/testify/assert"
)

func TestParseTime(t *testing.T) {
	html := `  <div class="cell small-2 text-center">
      <div class="time-primary">5:30 PM</div>
   </div>`
	result, err := ParseTime(html)
	assert.NoError(t, err)

	assert.Equal(t, "17:30", result)
}

type mockFetcher struct{}

func (m *mockFetcher) Fetch(url, class string) (string, error) {
	return "", nil
}

func TestParseSchedules(t *testing.T) {
	// Set a mock fetcher to avoid nil pointer dereference
	originalFetcher := VenueFetcher
	defer func() { VenueFetcher = originalFetcher }()
	VenueFetcher = &mockFetcher{}

	var cases = map[string]struct {
		file string
	}{
		"scta": {
			file: "../../testdata/scta.html",
		},
		"beechey": {
			file: "../../testdata/beechey.html",
		},
		"bluewaterhockey": {
			file: "../../testdata/bluewaterhockey.html",
		},
	}

	for site, tc := range cases {
		doc, err := htmlquery.LoadDoc(tc.file)
		assert.NoError(t, err)
		result := ParseSchedules(site, "", doc)

		// Verify basic structure: should have results with 7 fields each
		// [datetime, site, home_team, guest_team, location, division, address]
		assert.NotEmpty(t, result, "Expected non-empty results for site %s", site)
		for i, row := range result {
			assert.Equal(t, 7, len(row), "Expected 7 fields in row %d for site %s", i, site)
			assert.Equal(t, site, row[1], "Expected site field to match site name in row %d", i)
		}
	}
}

func TestGetVenueAddress(t *testing.T) {
	url := "https://www.theonedb.com/Venues/FindForGame/2067480"
	addr := GetVenueAddress(url, "remote")
	log.Println(addr)
	assert.NotEmpty(t, addr)
}
