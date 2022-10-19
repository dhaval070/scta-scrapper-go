package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"time"

	"flag"

	"calendar-scrapper/pkg/parser"
	"calendar-scrapper/pkg/writer"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

type ByDate [][]string

func (a ByDate) Len() int {
	return len(a)
}

func (a ByDate) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByDate) Less(i, j int) bool {
	return a[i][0] < a[j][0]
}

func parseGroups(doc *html.Node) map[string]string {
	links := htmlquery.Find(doc, `//div[@class="site-list"]/div/a`)

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

func fetchSchedules(url string, groups map[string]string, intdt int) [][]string {

	var schedules = make([][]string, 0)

	for division, id := range groups {
		url := fmt.Sprintf(url, id)
		doc, err := htmlquery.LoadURL(url)
		if err != nil {
			log.Fatal("load calendar url", err)
		}

		result := parser.ParseSchedules(doc, intdt)

		for i, row := range result {
			f := row[2:]
			result[i] = []string{row[0], division}
			result[i] = append(result[i], f...)
			schedules = append(schedules, result[i])
		}
	}

	sort.Sort(ByDate(schedules))
	return schedules
}

func main() {
	ymd := time.Now().Format("20060102")

	today := flag.String("today", ymd, "parse from date(yyyymmdd)")
	outfile := flag.String("outfile", "", "output filename")

	flag.Parse()
	log.Println(*today)

	var doc *html.Node
	var err error

	doc, err = htmlquery.LoadURL("https://etahockey.com/Seasons/Current/")

	intdt, err := strconv.Atoi(*today)
	if err != nil {
		log.Fatal(err)
	}

	groups := parseGroups(doc)

	var result = fetchSchedules("https://etahockey.com/Groups/%s/Calendar/", groups, intdt)

	if *outfile != "" {
		fh, err := os.Create(*outfile)
		if err != nil {
			log.Fatal(err)
		}
		writer.WriteCsv(fh, result)
	} else {
		log.Println(result)
	}
}
