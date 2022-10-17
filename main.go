//go:generate enumer -type=Month
package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

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
	reg := regexp.MustCompile("[0-9]{1,2}:[0-9]{1,2} (AM|PM)")

	return reg.FindString(content)
}

func QueryInnerText(doc *html.Node, expr string) string {

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
			division := QueryInnerText(item, `//div[@class="subject-group"]`)
			homeTeam := QueryInnerText(item, `//div[contains(@class, "subject-owner")]`)
			subjectText, err := htmlquery.Query(item, `//div[contains(@class, "subject-text")]`)
			if err != nil {
				log.Fatal(err)
			}

			ch := subjectText.FirstChild
			guestTeam := strings.Replace(htmlquery.InnerText(ch), "@ ", "", -1)
			location := QueryInnerText(item, `//div[@class="location remote"]`)
			log.Println(timeval, division, homeTeam, " vs ", guestTeam, " @ ", location)
			result = append(result, []string{ymd + " " + timeval, division, homeTeam, guestTeam, location})
		}
	}
	return result
}

func main() {
	doc, err := htmlquery.LoadDoc("testdata/scta.html")
	if err != nil {
		log.Fatal(err)
	}

	parseSchedules(doc, 20221017)
}
