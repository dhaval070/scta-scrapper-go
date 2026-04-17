package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"calendar-scrapper/config"
	"calendar-scrapper/dao/model"
	"calendar-scrapper/pkg/repository"
)

var repo *repository.Repository

// normalizeRowColumns ensures a CSV row has 7 columns by adding an empty event_id column if needed
// Input should be 6 columns: [datetime, site, home, guest, location, division]
// or 7 columns: [datetime, site, home, guest, location, division, event_id]
// Returns 7 columns: [datetime, site, home, guest, location, division, event_id]
func normalizeRowColumns(row []string) []string {
	switch len(row) {
	case 6:
		// Insert empty event_id at index 6
		return []string{row[0], row[1], row[2], row[3], row[4], row[5], ""}
	case 7:
		// Already has event_id column
		return row
	default:
		// Unexpected length, return as-is (will cause error downstream)
		return row
	}
}

func attachSurfaceID(site string, r io.Reader) [][]string {
	rr := csv.NewReader(r)

	var result = [][]string{}

	for {
		r, err := rr.Read()
		if errors.Is(err, io.EOF) {
			break
		}

		// Normalize to 7 columns (add empty event_id if missing)
		r = normalizeRowColumns(r)

		if len(r) != 7 {
			log.Fatalf("invalid columns after normalization %+v\n", r)
		}

		sl, err := repo.GetSitesLocation(site, r[4])
		if err != nil {
			log.Println(err)
			continue
		}

		r = append(r, strconv.FormatInt(int64(sl.LocationID), 10), strconv.FormatInt(int64(sl.SurfaceID), 10))

		result = append(result, r)
	}

	return result
}

func main() {
	config.Init("config", ".")

	var cfg = config.MustReadConfig()
	repo = repository.NewRepository(cfg)

	matchSurface := flag.Bool("match-surface", true, "import only surface matched rows")
	infile := flag.String("infile", "", "schedule csv file")
	site := flag.String("site", "", "site name")
	insert := flag.Bool("import", false, "--import")

	var sdate string
	flag.StringVar(&sdate, "cutoffdate", "", "-cutoffdate 2024-01-01")

	flag.Parse()

	if *insert && sdate == "" {
		log.Fatal("cutoff date is required to import")
	}

	var cdate time.Time
	var err error
	if *insert {
		cdate, err = time.Parse("2006-01-02", sdate)

		if err != nil {
			log.Fatal("failed to parse cutoff date", err)
		}
	}

	if *infile == "" {
		log.Fatal("infle is required")
	}

	if *site == "" {
		log.Fatal("site is required")
	}

	c, err := os.ReadFile(*infile)
	if err != nil {
		panic(err)
	}

	r := strings.NewReader(string(c))

	var result [][]string

	if *matchSurface {
		result = attachSurfaceID(*site, r)
	} else {
		rr := csv.NewReader(r)
		for {
			r, err := rr.Read()
			if errors.Is(err, io.EOF) {
				break
			}
			// Normalize to 7 columns (add empty event_id if missing)
			r = normalizeRowColumns(r)
			// add 0 surface ID
			r = append(r, "0", "0")
			result = append(result, r)
		}
	}

	err = repo.DB.Exec(`update sites_config set games_scraped=? where site_name=?`, len(result), *site).Error
	if err != nil {
		log.Printf("database error: failed to update site_scraped count for site %s , %s\n", *site, err.Error())
	}

	if *insert {
		log.Println("importing")
		if err = importEvents(repo, *site, result, cdate); err != nil {
			log.Println("failed to import ", err)
		}
	}

	ww := csv.NewWriter(os.Stdout)

	// Write rows, skipping column 6 (event_id) to maintain 8-column output format
	for _, row := range result {
		if len(row) == 9 {
			// Expected format: [datetime, site, home, guest, location, division, event_id, location_id, surface_id]
			// Output format: [datetime, site, home, guest, location, division, location_id, surface_id]
			outRow := []string{row[0], row[1], row[2], row[3], row[4], row[5], row[7], row[8]}
			if err := ww.Write(outRow); err != nil {
				panic(err)
			}
		} else {
			// Fallback for unexpected lengths (should not happen after normalization)
			if err := ww.Write(row); err != nil {
				panic(err)
			}
		}
	}

	ww.Flush()
}

func importEvents(repo *repository.Repository, site string, result [][]string, cutOffDate time.Time) error {
	var err error
	var SourceType = "scrape"
	// if gamesheet season
	if strings.HasPrefix(site, "gs_") {
		SourceType = ""
	}

	m := make([]*model.Event, 0, len(result))
	for _, rec := range result {
		// After normalization, rows should have 9 columns:
		// [datetime, site, home, guest, location, division, event_id, location_id, surface_id]
		if len(rec) != 9 {
			log.Printf("Warning: skipping row with unexpected column count %d: %v", len(rec), rec)
			continue
		}

		// Column indices are now fixed
		eventID := rec[6]
		locationId, err := strconv.Atoi(rec[7])
		if err != nil {
			return fmt.Errorf("failed to parse locationId %s, %w", rec[7], err)
		}
		sid, err := strconv.Atoi(rec[8])
		if err != nil {
			return fmt.Errorf("failed to parse surfaceId %s, %w", rec[8], err)
		}
		dt, err := time.Parse("2006-1-02 15:04", rec[0])
		if err != nil {
			return fmt.Errorf("failed to parse date %s, %w", rec[0], err)
		}

		if dt.Before(cutOffDate) {
			continue
		}

		m = append(m, &model.Event{
			Site:        rec[1],
			SourceType:  SourceType,
			Datetime:    dt,
			HomeTeam:    rec[2],
			GuestTeam:   rec[3],
			Location:    rec[4],
			Division:    rec[5],
			EventID:     eventID,
			LocationID:  int32(locationId),
			SurfaceID:   int32(sid),
			DateCreated: time.Now(),
		})
	}

	err = repo.ImportEvents(site, m, cutOffDate)
	return err
}
