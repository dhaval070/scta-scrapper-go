package main

import (
	"log"
	"regexp"
	"strconv"
	"time"

	"calendar-scrapper/pkg/cmdutil"
	"calendar-scrapper/pkg/htmlutil"
	"calendar-scrapper/pkg/parser"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

const SITE = "ucmhl"

func parseGroups(doc *html.Node) map[string]string {
	links := htmlquery.Find(doc, `//div[@class="site-list"]//a`)

	var groups = make(map[string]string)

	for _, n := range links {
		href := htmlutil.GetAttr(n, "href")
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

	doc, err = htmlquery.LoadURL("https://ucmhl.ca/Seasons/Current/")
	if err != nil {
		panic(err)
	}

	groups := parseGroups(doc)
	log.Println(groups)

	var result = parser.FetchSchedules(SITE, "https://ucmhl.ca/Groups/%s/Calendar/?Month=%d&Year=%d", groups, mm, yyyy)

	if *flags.ImportLocations {
		if err := cmdutil.ImportLocations(SITE, result); err != nil {
			log.Fatal(err)
		}
	}
	if err := cmdutil.WriteOutput(*flags.Outfile, result); err != nil {
		log.Fatal(err)
	}
}
func parseMonthYear(dt string) (int, int) {
	re := regexp.MustCompile(`^[0-9]{6}$`)

	if !re.Match([]byte(dt)) {
		panic("invalid format mmyyyy input")
	}

	mm, err := strconv.Atoi(dt[:2])
	if err != nil {
		panic(err)
	}

	yyyy, err := strconv.Atoi(dt[2:])
	if err != nil {
		panic(err)
	}
	return mm, yyyy
}
