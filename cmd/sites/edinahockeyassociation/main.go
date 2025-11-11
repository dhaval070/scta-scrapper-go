package main

import (
	"calendar-scrapper/pkg/parser"
	"calendar-scrapper/pkg/cmdutil"
	"time"

	"golang.org/x/net/html"
)

const SITE = "edinahockeyassociation"
const BASE_URL = "https://www.edinahockeyassociation.com/schedule/day/league_instance/216128/"

func main() {
	flags := cmdutil.ParseCommonFlags()

	var doc *html.Node
	var err error
	var mm, yyyy int

	if *flags.Date == "" {
		today := time.Now()
		mm = int(today.Month())
		yyyy = int(today.Year())

	} else {
		mm, yyyy = parser.ParseMonthYear(*flags.Date)
	}

}
