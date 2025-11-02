package main

import (
	"log"
	"time"

	"calendar-scrapper/pkg/parser"
	"calendar-scrapper/pkg/cmdutil"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

const SITE = "essexll"

func parseGroups(doc *html.Node) map[string]string {
	return parser.ParseSiteListGroups(doc, `//div[@class="site-list"]/div/a`)
}

func main() {
	flags := cmdutil.ParseCommonFlags()

	var doc *html.Node
	today := time.Now()
	mm := int(today.Month())
	yyyy := int(today.Year())

	if *flags.Date != "" {
		mm, yyyy = parser.ParseMonthYear(*flags.Date)
	}

	doc, err := htmlquery.LoadURL("https://essexll.ca/Seasons/Current/")
	if err != nil {
		log.Fatal(err)
	}

	groups := parseGroups(doc)
	log.Println(groups)

	var result = parser.FetchSchedules(SITE, "https://essexll.ca/Groups/%s/Calendar/?Month=%d&Year=%d", groups, mm, yyyy)

	if *flags.ImportLocations {
		if err := cmdutil.ImportLocations(SITE, result); err != nil {
			log.Fatal(err)
		}
	}
	if err := cmdutil.WriteOutput(*flags.Outfile, result); err != nil {
		log.Fatal(err)
	}
}
