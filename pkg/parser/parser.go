package parser

import (
	"calendar-scrapper/internal/client"
	"calendar-scrapper/pkg/month"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

var Client = client.GetClient(os.Getenv("HTTP_PROXY"))

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

func ParseTime(content string) (string, error) {
	reg := regexp.MustCompile("([0-9]{1,2}):([0-9]{1,2}) (AM|PM)")

	parts := reg.FindStringSubmatch(content)

	if parts == nil {
		return "", fmt.Errorf("failed to parse time %s", content)
	}

	h, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", errors.New("failed to convert hour")
	}

	m, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", errors.New("failed to convert minutes")
	}

	if parts[3] == "PM" && h < 12 {
		h += 12
	}

	return fmt.Sprintf("%02d:%02d", h, m), nil
}

func ParseSchedules(site string, doc *html.Node) [][]string {
	nodes := htmlquery.Find(doc, `//div[contains(@class, "day-details")]`)

	var result = [][]string{}

	var lock = &sync.Mutex{}
	var wg = &sync.WaitGroup{}

	for _, node := range nodes {
		id := GetAttr(node, "id")
		ymd := ParseId(id)

		listItems := htmlquery.Find(node, `//div[contains(@class, "event-list-item")]/div`) // `div[2]`)

		for _, parent := range listItems {
			item := htmlquery.FindOne(parent, `div[2]`)
			content := htmlquery.OutputHTML(item, true)

			if strings.Contains(content, "All Day") || strings.Contains(content, "time-secondary") || strings.Contains(content, "Cancelled") {
				continue
			}
			timeval, err := ParseTime(content)
			if err != nil {
				log.Fatal(err, content)
				continue
			}

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

			item = htmlquery.Find(parent, `div[1]//a[@class="remote" or @class="local"]`)[0]
			var url string
			var address string
			var class string

			for _, attr := range item.Attr {
				if attr.Key == "href" {
					url = attr.Val
					break
				} else if attr.Key == "class" {
					class = attr.Val
				}
			}

			if url != "" {
				wg.Add(1)
				go func(url string, wg *sync.WaitGroup, lock *sync.Mutex) {
					defer wg.Done()
					address = GetVenueAddress(url, class)
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

func ParseId(id string) string {
	parts := strings.Split(id, "-")

	if len(parts) != 4 {
		log.Fatal("len not 4")
	}
	mm, err := month.MonthString(parts[1])

	if err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%s-%02d-%s", parts[3], mm, parts[2])
}

func GetVenueAddress(url string, class string) string {
	req, err := http.NewRequest("GET", url, nil)
	resp, err := Client.Do(req)

	if err != nil {
		log.Println(err)
		return ""
	}
	defer resp.Body.Close()

	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		log.Println("error getting "+url, err)
		return ""
	}

	var address string

	// theonedb.com
	if class == "remote" {
		item := htmlquery.FindOne(doc, `//div[@class="container"]/div/div/h2/small[2]`)
		if item == nil {
			log.Println("address node not found, url:", url)
			return ""
		}
		address = htmlquery.InnerText(item)
	} else if class == "local" {
		// local url
		node := htmlquery.FindOne(doc, `//div[@class="month"]/following-sibling::div/div/div`)
		if node == nil {
			log.Println("address node not found, url:", url)
			return ""
		}
		address = htmlquery.InnerText(node)
	}
	log.Println(url + ":" + address)
	return address
}

func ParseMonthYear(dt string) (int, int) {
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

func FetchSchedules(site, url string, groups map[string]string, mm, yyyy int) [][]string {

	var schedules = make([][]string, 0)

	for division, id := range groups {
		url := fmt.Sprintf(url, id, mm, yyyy)
		doc, err := htmlquery.LoadURL(url)
		if err != nil {
			log.Fatal("load calendar url", err)
		}

		result := ParseSchedules(site, doc)

		for _, row := range result {
			row[5] = division
			schedules = append(schedules, row)
		}
	}

	sort.Sort(ByDate(schedules))
	return schedules
}
