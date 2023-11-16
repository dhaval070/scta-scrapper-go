package parser

import (
	"calendar-scrapper/pkg/month"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"

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

func ParseSchedules(site string, doc *html.Node, today int) [][]string {
	nodes := htmlquery.Find(doc, `//div[contains(@class, "day-details")]`)

	var result = [][]string{}

	var lock = &sync.Mutex{}
	var wg = &sync.WaitGroup{}

	for _, node := range nodes {
		id := GetAttr(node, "id")
		dt, ymd := ParseId(id)

		if dt < today {
			continue
		}
		log.Println("dt: ", dt)
		listItems := htmlquery.Find(node, `//div[contains(@class, "event-list-item")]/div`) // `div[2]`)

		for _, parent := range listItems {
			items := htmlquery.Find(parent, `div[2]`)
			item := items[0]
			content := htmlquery.OutputHTML(item, true)

			if strings.Contains(content, "All Day") || strings.Contains(content, "time-secondary") || strings.Contains(content, "Cancelled") {
				continue
			}
			timeval := ParseTime(content)
			division, err := QueryInnerText(item, `//span[@class="game_no"]`)
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
			if err != nil {
				log.Println("failed to parse location")
				continue
			}

			item = htmlquery.Find(parent, `div[1]//a[@class="remote"]`)[0]
			var url string
			var address string

			for _, attr := range item.Attr {
				if attr.Key == "href" {
					url = attr.Val
					break
				}
			}

			if url != "" {
				wg.Add(1)
				go func(url string, wg *sync.WaitGroup, lock *sync.Mutex) {
					defer wg.Done()
					address = GetVenueAddress(url)
					lock.Lock()
					result = append(result, []string{ymd + " " + timeval, site, homeTeam, guestTeam, location, division, address})
					lock.Unlock()
				}(url, wg, lock)
			} else {
				result = append(result, []string{ymd + " " + timeval, site, homeTeam, guestTeam, location, division, address})
			}
		}
	}
	wg.Wait()
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

func GetVenueAddress(url string) string {
	doc, err := htmlquery.LoadURL(url)
	if err != nil {
		log.Println("error getting "+url, err)
		return ""
	}

	item := htmlquery.FindOne(doc, `//div[@class="container"]/div/div/h2/small[2]`)
	if item == nil {
		log.Println("address node not found")
		return ""
	}
	address := htmlquery.InnerText(item)
	log.Println(url + ":" + address)
	return address
}
