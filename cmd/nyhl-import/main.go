package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"calendar-scrapper/config"
	"calendar-scrapper/dao/model"
	"calendar-scrapper/pkg/repository"
)

var repo *repository.Repository

func main() {
	config.Init("config", ".")

	var cfg = config.MustReadConfig()
	repo = repository.NewRepository(cfg)

	infile := flag.String("infile", "", "schedule csv file")

	var sdate string
	flag.StringVar(&sdate, "cutoffdate", "", "-cutoffdate 2024-01-01")

	flag.Parse()

	if sdate == "" {
		log.Fatal("cutoff date is required to import")
	}

	var cdate time.Time
	var err error
	cdate, err = time.Parse("2006-01-02", sdate)

	if err != nil {
		log.Fatal("failed to parse cutoff date", err)
	}

	if *infile == "" {
		log.Fatal("infle is required")
	}

	f, err := os.Open(*infile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	m, err := repo.GetNyhlMappings()
	if err != nil {
		log.Fatal(err)
	}
	err = importEvents(r, cdate, m)
	if err != nil {
		log.Fatal(err)
	}
}

// delete the two group columns before import
//format required: GameID	League	Season	Division	Tier	group HomeTeam	Tier group	VisitorTeam	Location	Date	Time
func importEvents(ff *csv.Reader, cutOffDate time.Time, mapping map[string]int) error {
	var err error
	var SourceType = "file"

	m := []*model.Event{}

	for i := 1; ; i += 1 {
		cols, err := ff.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(fmt.Errorf("error at row %d, %w", i, err))
		}

		sid, ok := mapping[cols[10]]
		if !ok {
			log.Printf("failed to map surfaceid %s\n", cols[10])
		}

		dt, err := parseDate(cols[11], cols[12])
		if err != nil {
			return err
		}

		if dt.Before(cutOffDate) {
			continue
		}

		m = append(m, &model.Event{
			Site:        "nyhl",
			SourceType:  SourceType,
			Datetime:    dt,
			HomeTeam:    cols[6],
			GuestTeam:   cols[9],
			Location:    cols[10],
			Division:    cols[3],
			SurfaceID:   int32(sid),
			DateCreated: time.Now(),
		})
	}

	log.Println("total events ", len(m))
	err = repo.ImportEvents("nyhl", m, cutOffDate)
	return err
}

func parseDate(date, t string) (tt time.Time, err error) {
	t1, err := time.Parse("3:04 PM", t)
	if err != nil {
		return tt, fmt.Errorf("failed to parse time:%s %w", t, err)
	}

	dt, err := time.Parse("1/2/2006 15:04:05", date+" "+t1.Format("15:04:05"))
	if err != nil {
		return tt, fmt.Errorf("failed to parse date:%s time:%s %w", date, t, err)
	}
	return dt, nil
}
