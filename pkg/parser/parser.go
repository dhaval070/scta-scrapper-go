package parser

import (
	"calendar-scrapper/pkg/month"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

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

func GetAttr(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func ParseTime(content string) string {
	reg := regexp.MustCompile("([0-9]{1,2}):([0-9]{1,2}) (AM|PM)")

	parts := reg.FindStringSubmatch(content)

	if parts == nil {
		log.Fatal("failed to parse time: " + content)
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

func ParseSchedules(doc *html.Node, today int) [][]string {
	nodes := htmlquery.Find(doc, `//div[contains(@class, "day-details")]`)

	var result = [][]string{}

	for _, node := range nodes {
		id := GetAttr(node, "id")
		dt, ymd := ParseId(id)

		if dt < today {
			continue
		}
		log.Println("dt: ", dt)
		listItems := htmlquery.Find(node, `//div[contains(@class, "event-list-item")]/div/div[2]`)

		for _, item := range listItems {
			content := htmlquery.OutputHTML(item, true)

			if strings.Contains(content, "All Day") || strings.Contains(content, "time-secondary") || strings.Contains(content, "Cancelled") {
				continue
			}
			timeval := ParseTime(content)
			division, err := QueryInnerText(item, `//div[@class="subject-group"]`)
			if err != nil {
				log.Println(err)
				continue
			}
			guestTeam, err := QueryInnerText(item, `//div[contains(@class, "subject-owner")]`)
			if err != nil {
				log.Println(err)
				continue
			}
			subjectText, err := htmlquery.Query(item, `//div[contains(@class, "subject-text")]`)

			if err != nil {
				log.Println(err)
				continue
			}

			ch := subjectText.FirstChild
			homeTeam := strings.Replace(htmlquery.InnerText(ch), "@ ", "", -1)
			location, err := QueryInnerText(item, `//div[@class="location remote"]`)
			result = append(result, []string{ymd + " " + timeval, "", homeTeam, guestTeam, location, division})
		}
	}
	return result
}

func QueryInnerText(doc *html.Node, expr string) (string, error) {
	node, err := htmlquery.Query(doc, expr)
	if err != nil {
		return "", err
	}
	if node != nil {
		return htmlquery.InnerText(node), nil
	}
	return "", fmt.Errorf("node not found %v", expr)
}

func ParseId(id string) (int, string) {
	parts := strings.Split(id, "-")

	if len(parts) != 4 {
		log.Fatal("len not 4")
	}
	mm, err := month.MonthString(parts[1])

	if err != nil {
		log.Fatal(err)
	}

	dt, err := strconv.Atoi(fmt.Sprintf("%s%02d%s", parts[3], mm, parts[2]))
	if err != nil {
		log.Fatal(err)
	}

	return dt, fmt.Sprintf("%s-%02d-%s", parts[3], mm, parts[2])

}
