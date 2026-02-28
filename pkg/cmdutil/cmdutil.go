package cmdutil

import (
	"calendar-scrapper/config"
	"calendar-scrapper/dao/model"
	"calendar-scrapper/pkg/repository"
	"calendar-scrapper/pkg/writer"
	"flag"
	"log"
	"os"
)

// Flags holds common command-line flags for site scrapers
type Flags struct {
	Date            *string
	Outfile         *string
	ImportLocations *bool
}

// ParseCommonFlags defines and parses common flags used by most site scrapers
func ParseCommonFlags() *Flags {
	f := &Flags{
		Date:            flag.String("date", "", "calendar month and year in format: mmyyyy"),
		Outfile:         flag.String("outfile", "", "output filename"),
		ImportLocations: flag.Bool("import-locations", false, "import site locations"),
	}
	flag.Parse()
	return f
}

// ImportLocations imports location data to the database
// This is used by virtually all site scrapers (72+ sites)
// result is expected to be [][]string where each row has:
// [date, site, home, guest, location, division, address]
func ImportLocations(siteName string, result [][]string) error {
	// Skip special importers - they use league-specific mapping tables (gthl_mappings, nyhl_mappings, mhl_mappings)
	// and lack address data, making location import unnecessary
	specialImporters := map[string]bool{
		"gthl": true,
		"nyhl": true,
		"mhl":  true,
	}
	if specialImporters[siteName] {
		log.Printf("Skipping location import for %s (uses league-specific mapping table)", siteName)
		return nil
	}

	config.Init("config", ".")
	cfg := config.MustReadConfig()

	var locations = make([]model.SitesLocation, 0, len(result))
	for _, r := range result {
		log.Printf("%+v\n", r)

		l := model.SitesLocation{
			Location: r[4],
			Address:  r[6],
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
