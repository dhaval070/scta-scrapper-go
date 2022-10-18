//go:generate enumer -type=Month
package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"flag"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

func getAttr(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

type Month int

const (
	Jan Month = iota + 1
	Feb
	Mar
	Apr
	May
	Jun
	Jul
	Aug
	Sep
	Oct
	Nov
	Dec
)

func parseId(id string) (int, string) {
	parts := strings.Split(id, "-")

	if len(parts) != 4 {
		log.Fatal("len not 4")
	}
	mm, err := MonthString(parts[1])

	if err != nil {
		log.Fatal(err)
	}

	dt, err := strconv.Atoi(fmt.Sprintf("%s%02d%s", parts[3], mm, parts[2]))
	if err != nil {
		log.Fatal(err)
	}

	return dt, fmt.Sprintf("%s-%02d-%s", parts[3], mm, parts[2])

}

func parseTime(content string) string {
	reg := regexp.MustCompile("([0-9]{1,2}):([0-9]{1,2}) (AM|PM)")

	parts := reg.FindStringSubmatch(content)

	if parts == nil {
		log.Fatal("failed to parse time")
	}

	h, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Fatal("failed to convert hour")
	}

	m, err := strconv.Atoi(parts[2])
	if err != nil {
		log.Fatal("failed to convert minutes")
	}

	if parts[3] == "PM" && h < 12 {
		h += 12
	}

	return fmt.Sprintf("%02d:%02d", h, m)
}

func queryInnerText(doc *html.Node, expr string) string {

	node, err := htmlquery.Query(doc, expr)
	if err != nil {
		log.Fatal(err)
	}
	if node != nil {
		return htmlquery.InnerText(node)
	}
	return ""
}

func parseSchedules(doc *html.Node, today int) [][]string {
	nodes := htmlquery.Find(doc, `//div[contains(@class, "day-details")]`)

	var result = [][]string{}

	for _, node := range nodes {
		id := getAttr(node, "id")
		dt, ymd := parseId(id)

		if dt < today {
			continue
		}
		log.Println("dt: ", dt)
		listItems := htmlquery.Find(node, `//div[contains(@class, "event-list-item")]/div/div[2]`)

		for _, item := range listItems {
			content := htmlquery.OutputHTML(item, true)
			timeval := parseTime(content)
			division := queryInnerText(item, `//div[@class="subject-group"]`)
			guestTeam := queryInnerText(item, `//div[contains(@class, "subject-owner")]`)
			subjectText, err := htmlquery.Query(item, `//div[contains(@class, "subject-text")]`)
			if err != nil {
				log.Fatal(err)
			}

			ch := subjectText.FirstChild
			homeTeam := strings.Replace(htmlquery.InnerText(ch), "@ ", "", -1)
			location := queryInnerText(item, `//div[@class="location remote"]`)
			result = append(result, []string{ymd + " " + timeval, division, homeTeam, guestTeam, location})
		}
	}
	return result
}

func writeCsv(filename string, data [][]string) {
	fh, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}

	csv := csv.NewWriter(fh)

	csv.Write([]string{"date", "division", "home team", "guest team", "location"})

	for _, row := range data {
		csv.Write(row)
	}
	fh.Close()
}

func main() {
	ymd := time.Now().Format("20060102")

	infile := flag.String("infile", "", "local html filename")
	today := flag.String("today", ymd, "parse from date(yyyymmdd)")
	outfile := flag.String("outfile", "", "output filename")

	flag.Parse()

	log.Println(*today)
	log.Println(*infile)

	var doc *html.Node
	var err error

	if *infile != "" {
		doc, err = htmlquery.LoadDoc("testdata/scta.html")
	} else {
		doc, err = htmlquery.LoadURL("https://sctahockey.com/Calendar/")
	}

	if err != nil {
		log.Fatal(err)
	}

	intdt, err := strconv.Atoi(*today)
	if err != nil {
		log.Fatal(err)
	}

	result := parseSchedules(doc, intdt)

	if *outfile != "" {
		writeCsv(*outfile, result)
	} else {
		log.Println(result)
	}

}
