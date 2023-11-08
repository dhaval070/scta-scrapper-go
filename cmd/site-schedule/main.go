package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"calendar-scrapper/config"
	"calendar-scrapper/pkg/repository"
)

var repo *repository.Repository

func processCsv(site string, r io.Reader) [][]string {
	rr := csv.NewReader(r)

	var result = [][]string{}

	for {
		r, err := rr.Read()
		if errors.Is(err, io.EOF) {
			break
		}

		if len(r) != 7 {
			log.Fatalf("invalid columns %+v\n", r)
		}

		sl, err := repo.GetSitesLocation(site, r[4])
		if err != nil {
			log.Println(err)
			continue
		}
		if sl.LocationID != 0 {
			// log.Println(sl.Location, sl.LocationID)
			r = append(r, fmt.Sprint(sl.LocationID))
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

	infile := flag.String("infile", "", "schedule csv file")
	site := flag.String("site", "", "site name")
	flag.Parse()

	if infile == nil {
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
	// fh, err := os.Open(*infile)
	// if err != nil {
	// 	panic(err)
	// }

	result := processCsv(*site, r)

	ww := csv.NewWriter(os.Stdout)

	err = ww.WriteAll(result)
	if err != nil {
		panic(err)
	}

	ww.Flush()
}
