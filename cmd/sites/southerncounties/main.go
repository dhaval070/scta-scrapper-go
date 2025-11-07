package main

import (
	"log"
	"regexp"
	"time"

	"calendar-scrapper/pkg/cmdutil"
	"calendar-scrapper/pkg/htmlutil"
	"calendar-scrapper/pkg/parser"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

const SITE = "southerncounties"

func parseGroups(doc *html.Node) map[string]string {
	links := htmlquery.Find(doc, `//div[contains(@class, "site-list")]/a`)

	var groups = make(map[string]string)

	for _, n := range links {
		href := htmlutil.GetAttr(n, "href")
		division := htmlquery.InnerText(n)

		re := regexp.MustCompile(`Groups/(.+)/`)

		parts := re.FindAllStringSubmatch(href, -1)
		if parts == nil {
			continue
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

	doc, err = htmlquery.LoadURL("https://southerncounties.ca/Seasons/Current/")
	// fmt.Println(htmlquery.OutputHTML(doc, true))

	groups := parseGroups(doc)
	log.Println(groups)

	var result = parser.FetchSchedules(SITE, "https://southerncounties.ca/Groups/%s/Calendar/?Month=%d&Year=%d", groups, mm, yyyy)

	if *flags.ImportLocations {
		if err := cmdutil.ImportLocations(SITE, result); err != nil {
			log.Fatal(err)
		}
	}
	if err := cmdutil.WriteOutput(*flags.Outfile, result); err != nil {
		log.Fatal(err)
	}
}
