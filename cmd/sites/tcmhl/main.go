package main

import (
	"log"
	"regexp"
	"time"

	"calendar-scrapper/pkg/parser"
	"calendar-scrapper/pkg/cmdutil"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

const SITE = "tcmhl"

func parseGroups(doc *html.Node) map[string]string {
	links := htmlquery.Find(doc, `//div[@class="site-list"]/div/div/div/div/div/a`)

	var groups = make(map[string]string)

	for _, n := range links {
		href := parser.GetAttr(n, "href")
		division := htmlquery.InnerText(n)

		re := regexp.MustCompile(`Groups/(.+)/`)

		parts := re.FindAllStringSubmatch(href, -1)
		if parts == nil {
			log.Fatal("failed to parse group link", href)
		}
		groups[division] = parts[0][1]
	}
	return groups
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

	doc, err = htmlquery.LoadURL("https://tcmhl.ca/Seasons/Current/")
	if err != nil {
		log.Fatal(err)
	}

	groups := parseGroups(doc)
	log.Println(groups)

	var result = parser.FetchSchedules(SITE, "https://tcmhl.ca/Groups/%s/Calendar/?Month=%d&Year=%d", groups, mm, yyyy)

	if *flags.ImportLocations {
		if err := cmdutil.ImportLocations(SITE, result); err != nil {
			log.Fatal(err)
		}
	}
	if err := cmdutil.WriteOutput(*flags.Outfile, result); err != nil {
		log.Fatal(err)
	}
}
