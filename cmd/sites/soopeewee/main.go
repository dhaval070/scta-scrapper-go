package main

import (
	"fmt"
	"log"
	"time"

	"calendar-scrapper/internal/soopeewee"
	"calendar-scrapper/pkg/parser"
	"calendar-scrapper/pkg/cmdutil"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

const SITE = "soopeewee"
const BASE_URL = "https://soopeewee.ca/"

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

	doc, err = htmlquery.LoadURL(fmt.Sprintf(BASE_URL+"Calendar/?Month=%d&Year=%d", mm, yyyy))
	if err != nil {
		panic(err)
	}

	var result = soopeewee.ParseSchedules(doc, SITE, BASE_URL)

	if *flags.ImportLocations {
		if err := cmdutil.ImportLocations(SITE, result); err != nil {
			log.Fatal(err)
		}
	}
	if err := cmdutil.WriteOutput(*flags.Outfile, result); err != nil {
		log.Fatal(err)
	}
}
