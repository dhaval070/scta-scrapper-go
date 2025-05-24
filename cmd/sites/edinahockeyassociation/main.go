package main

import (
	"calendar-scrapper/pkg/parser"
	"flag"
	"time"

	"golang.org/x/net/html"
)

const SITE = "edinahockeyassociation"
const BASE_URL = "https://www.edinahockeyassociation.com/schedule/day/league_instance/216128/"

func main() {
	date := flag.String("date", "", "calendar month and year in format: mmyyyy")
	outfile := flag.String("outfile", "", "output filename")
	importLocations := flag.Bool("import-locations", false, "import site locations")

	flag.Parse()

	var doc *html.Node
	var err error
	var mm, yyyy int

	if *date == "" {
		today := time.Now()
		mm = int(today.Month())
		yyyy = int(today.Year())

	} else {
		mm, yyyy = parser.ParseMonthYear(*date)
	}

}
