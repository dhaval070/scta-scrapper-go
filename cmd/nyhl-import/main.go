package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"time"

	"calendar-scrapper/config"
	"calendar-scrapper/dao/model"
	"calendar-scrapper/pkg/repository"

	"github.com/thedatashed/xlsxreader"
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

	xl, err := xlsxreader.OpenFile(*infile)
	if err != nil {
		log.Fatal(err)
	}
	err = importEvents(xl, cdate)
	if err != nil {
		log.Fatal(err)
	}
}

// delete the two group columns before import
//format required: GameID	League	Season	Division	Tier	Home Team	Tier	Visitor Team	Location	Date	Time	LB Surface ID
func importEvents(xl *xlsxreader.XlsxFileCloser, cutOffDate time.Time) error {
	var err error
	var SourceType = "file"

	m := []*model.Event{}

	ch := xl.ReadRows(xl.Sheets[0])

	for rec := range ch {
		if rec.Error != nil {
			return err
		}

		if rec.Cells[11].Type != xlsxreader.TypeNumerical {
			return fmt.Errorf("invalid type for surface id %v %s", rec.Cells[11].Type, rec.Cells[13].Value)
		}
		sid, err := strconv.Atoi(rec.Cells[11].Value)
		if err != nil {
			return fmt.Errorf("failed to parse surfaceid %s", rec.Cells[11].Value)
		}

		dt, err := parseDate(rec.Cells[9].Value, rec.Cells[10].Value)
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
			HomeTeam:    rec.Cells[6].Value,
			GuestTeam:   rec.Cells[8].Value,
			Location:    rec.Cells[8].Value,
			Division:    rec.Cells[3].Value,
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
