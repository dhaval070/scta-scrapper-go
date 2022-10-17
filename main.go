//go:generate enumer -type=Month
package main

import (
	"bytes"
	"fmt"
	"log"
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

func parseId(id string) int {
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

	return dt
}

func parseSchedules(doc *html.Node, today int) {
	nodes := htmlquery.Find(doc, `//div[contains(@class, "day-details")]`)

	for _, node := range nodes {
		id := getAttr(node, "id")

		dt := parseId(id)

		if dt < today {
			continue
		}
		log.Println("dt: ", dt)
		listItems := htmlquery.Find(node, `//div[contains(@class, "event-list-item")]/div/div[2]`)

		for _, item := range listItems {
			writer := bytes.NewBufferString("")
			html.Render(writer, item)

			log.Println(writer.String())

		}

	}
}

func main() {
	doc, err := htmlquery.LoadDoc("testdata/scta.html")
	if err != nil {
		log.Fatal(err)
	}

	parseSchedules(doc, 20221017)
}
