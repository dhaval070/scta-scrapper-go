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

func attachSurfaceID(site string, r io.Reader) [][]string {
	rr := csv.NewReader(r)

	var result = [][]string{}

	for {
		r, err := rr.Read()
		if errors.Is(err, io.EOF) {
			break
		}

		if len(r) != 6 {
			log.Fatalf("invalid columns %+v\n", r)
		}

		sl, err := repo.GetSitesLocation(site, r[4])
		if err != nil {
			log.Println(err)
			continue
		}
		if sl.SurfaceID != 0 {
			// log.Println(sl.Location, sl.LocationID)
			r = append(r, fmt.Sprint(sl.SurfaceID))
			result = append(result, r)
		} else {
			log.Println("skipped site: ", site, ", location: ", sl.Location)
		}
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
			// add 0 surface ID
			r = append(r, "0")
			result = append(result, r)
		}
	}
	if *insert {
		log.Println("importing")
		if err = importEvents(repo, *site, result, cdate); err != nil {
			log.Println("failed to import ", err)
		}
	}

	ww := csv.NewWriter(os.Stdout)

	err = ww.WriteAll(result)
	if err != nil {
		panic(err)
	}

	ww.Flush()
}

func importEvents(repo *repository.Repository, site string, result [][]string, cutOffDate time.Time) error {
	var err error
	var SourceType = "scrape"

	m := make([]*model.Event, 0, len(result))
	for _, rec := range result {
		sid, err := strconv.Atoi(rec[6])
		if err != nil {
			return fmt.Errorf("failed to parse surfaceid %s, %w", rec[6], err)
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
			SurfaceID:   int32(sid),
			DateCreated: time.Now(),
		})
	}

	err = repo.ImportEvents(site, m, cutOffDate)
	return err
}
