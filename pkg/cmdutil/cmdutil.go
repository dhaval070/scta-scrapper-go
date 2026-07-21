package cmdutil

import (
	"calendar-scrapper/config"
	"calendar-scrapper/dao/model"
	"calendar-scrapper/pkg/repository"
	"calendar-scrapper/pkg/writer"
	"flag"
	"log"
	"os"
	"strings"
	"time"
)

// SpecialImporters are sites that use league-specific mapping tables and already import events via their own binaries
var SpecialImporters = map[string]bool{
	"gthl": true,
	"nyhl": true,
	"mhl":  true,
}

// Flags holds common command-line flags for site scrapers
type Flags struct {
	Date            *string
	Outfile         *string
	ImportLocations *bool
	ImportEvents    *bool
	CutoffDate      *string
}

// ParseCommonFlags defines and parses common flags used by most site scrapers
func ParseCommonFlags() *Flags {
	f := &Flags{
		Date:            flag.String("date", "", "calendar month and year in format: mmyyyy"),
		Outfile:         flag.String("outfile", "", "output filename"),
		ImportLocations: flag.Bool("import-locations", false, "import site locations"),
		ImportEvents:    flag.Bool("import-events", false, "import events to database"),
		CutoffDate:      flag.String("cutoff-date", "", "cutoff date for event import (YYYY-MM-DD)"),
	}
	flag.Parse()
	return f
}

// normalizeCSVRows ensures all CSV rows have 8 columns by adding empty event_id if needed
// Input: 7 columns [datetime, site, home, guest, location, division, address]
//
//	8 columns [datetime, site, home, guest, location, division, event_id, address]
//
// Output: 8 columns [datetime, site, home, guest, location, division, event_id, address]
func normalizeCSVRows(rows [][]string) [][]string {
	normalized := make([][]string, 0, len(rows))
	for _, row := range rows {
		switch len(row) {
		case 7:
			// Add empty event_id at index 6, shift address to index 7
			normalized = append(normalized, []string{
				row[0], row[1], row[2], row[3], row[4], row[5], "", row[6],
			})
		case 8:
			// Already has event_id
			normalized = append(normalized, row)
		default:
			// Skip invalid rows
			log.Printf("Warning: skipping row with invalid column count %d: %v", len(row), row)
		}
	}
	return normalized
}

// ImportLocations imports location data to the database
// This is used by virtually all site scrapers (72+ sites)
// result is expected to be [][]string where each row has:
// 8 columns: [date, site, home, guest, location, division, event_id, address] (after normalization)
func ImportLocations(siteName string, result [][]string) error {
	// Skip special importers - they use league-specific mapping tables (gthl_mappings, nyhl_mappings, mhl_mappings)
	// and lack address data, making location import unnecessary
	if SpecialImporters[siteName] {
		log.Printf("Skipping location import for %s (uses league-specific mapping table)", siteName)
		return nil
	}

	// Normalize all rows to 8 columns
	result = normalizeCSVRows(result)
	if len(result) == 0 {
		log.Printf("No valid rows to import locations for %s", siteName)
		return nil
	}

	config.Init("config", ".")
	cfg := config.MustReadConfig()

	var locations = make([]model.SitesLocation, 0, len(result))
	for _, r := range result {
		log.Printf("%+v\n", r)

		// After normalization, all rows have 8 columns
		// address is always at index 7
		l := model.SitesLocation{
			Location: r[4],
			Address:  r[7],
		}
		locations = append(locations, l)
	}

	repo := repository.NewRepository(cfg).Site(siteName)
	if err := repo.ImportLoc(locations); err != nil {
		return err
	}
	return nil
}

// WriteOutput writes the result to CSV file or logs it
// Used by 76+ sites with identical code
func WriteOutput(outfile string, result [][]string) error {
	if outfile != "" {
		var fh *os.File
		var err error

		if outfile == "-" {
			fh = os.Stdout
		} else {
			fh, err = os.Create(outfile)
			if err != nil {
				return err
			}
			defer fh.Close()
		}
		return writer.WriteCsv(fh, result)
	} else {
		log.Println(result)
	}
	return nil
}

// ImportEventsFromRows imports events from scraped rows to the database
// Rows are expected to have 8 columns: [datetime, site, home, guest, location, division, event_id, address] (after normalization)
func ImportEventsFromRows(repo *repository.Repository, site string, rows [][]string, cutoffDate time.Time) error {
	// Skip special importers - they already import events via their own binaries
	if SpecialImporters[site] {
		log.Printf("Skipping event import for %s (uses league-specific mapping table)", site)
		return nil
	}

	// Update games_scraped count (total rows before filtering)
	if err := repo.DB.Exec(`UPDATE sites_config SET games_scraped = ? WHERE site_name = ?`, len(rows), site).Error; err != nil {
		log.Printf("database error: failed to update games_scraped count for site %s: %v", site, err)
	}

	// Normalize all rows to 8 columns
	rows = normalizeCSVRows(rows)
	if len(rows) == 0 {
		log.Printf("No valid rows to import events for %s", site)
		return nil
	}

	var events []*model.Event
	for _, row := range rows {
		// After normalization, all rows have 8 columns
		// event_id is at index 6, address at index 7
		eventID := row[6]

		// Parse datetime
		dt, err := time.Parse("2006-1-02 15:04", row[0])
		if err != nil {
			log.Printf("Failed to parse datetime %s: %v", row[0], err)
			continue
		}

		// Skip rows before cutoff date
		if dt.Before(cutoffDate) {
			continue
		}

		// Get location mapping
		siteLoc, err := repo.GetSitesLocation(site, row[4])
		if err != nil {
			log.Printf("No mapping found for site %s location %s: %v", site, row[4], err)
			continue
		}

		// Determine source type
		sourceType := "scrape"
		if strings.HasPrefix(site, "gs_") {
			sourceType = ""
		}

		events = append(events, &model.Event{
			Site:        site,
			SourceType:  sourceType,
			Datetime:    dt,
			HomeTeam:    row[2],
			GuestTeam:   row[3],
			Location:    row[4],
			Division:    row[5],
			EventID:     eventID,
			LocationID:  siteLoc.LocationID,
			SurfaceID:   siteLoc.SurfaceID,
			DateCreated: time.Now(),
		})
	}

	if len(events) == 0 {
		log.Printf("No events to import for %s after filtering", site)
		return nil
	}

	log.Printf("Importing %d events for %s", len(events), site)
	return repo.ImportEvents(site, events, cutoffDate)
}
