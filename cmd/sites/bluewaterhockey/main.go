package main

import (
	"fmt"
	"log"
	"sort"
	"time"

	"calendar-scrapper/pkg/parser"
	"calendar-scrapper/pkg/cmdutil"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

const SITE = "bluewaterhockey"

func parseGroups(doc *html.Node) map[string]string {
	return parser.ParseSiteListGroups(doc, `//div[@class="site-list"]/div/div/div/a`)
}

func fetchSchedules(url string, groups map[string]string, mm, yyyy int) [][]string {

	var schedules = make([][]string, 0)

	for division, id := range groups {
		url := fmt.Sprintf(url, id, mm, yyyy)
		doc, err := htmlquery.LoadURL(url)
		if err != nil {
			log.Fatal("load calendar url", err)
		}

		result := parser.ParseSchedules(SITE, doc)

		for _, row := range result {
			row[5] = division
			schedules = append(schedules, row)
		}
	}

	sort.Sort(parser.ByDate(schedules))
	return schedules
}

func main() {
	flags := cmdutil.ParseCommonFlags()

	var doc *html.Node
	var err error
	today := time.Now()
	mm := int(today.Month())
	yyyy := int(today.Year())

	if *flags.Date != "" {
		mm, yyyy = parser.ParseMonthYear(*flags.Date)
	}

	doc, err = htmlquery.LoadURL("https://bluewaterhockey.ca/Seasons/Current/")

	groups := parseGroups(doc)
	log.Println(groups)

	var result = parser.FetchSchedules(SITE, "https://bluewaterhockey.ca/", "https://bluewaterhockey.ca/Groups/%s/Calendar/?Month=%d&Year=%d", groups, mm, yyyy)
	if *flags.ImportLocations {
		if err := cmdutil.ImportLocations(SITE, result); err != nil {
			log.Fatal(err)
		}
	}
	if err := cmdutil.WriteOutput(*flags.Outfile, result); err != nil {
		log.Fatal(err)
	}
}
