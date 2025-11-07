package soopeewee

import (
	"calendar-scrapper/pkg/htmlutil"
	"calendar-scrapper/pkg/parser"
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

func ParseSchedules(doc *html.Node, Site, baseURL string) [][]string {
	nodes := htmlquery.Find(doc, `//div[contains(@class, "day-details")]`)

	var result = [][]string{}

	var lock = &sync.Mutex{}
	var wg = &sync.WaitGroup{}

	// https://soopeewee.ca has three types of games home, local and away.
	reHome := regexp.MustCompile("(?i)(home|local) game")
	reAway := regexp.MustCompile("(?i)away game")

	for _, node := range nodes {
		listItems := htmlquery.Find(node, `//div[contains(@class, "event-list-item")]/div`)

		var id = htmlutil.GetAttr(node, "id")

		if id == "" {
			log.Fatal("id not found")
		}
		ymd := parser.ParseId(id)

		for _, parent := range listItems {
			item := htmlquery.FindOne(parent, `div[2]`)
			content := htmlquery.OutputHTML(item, true)

			var homeGame bool

			if reHome.MatchString(content) {
				homeGame = true
			} else if !reAway.MatchString(content) {
				// neither home game or away game then skip
				continue
			}

			timeval, err := parser.ParseTime(content)
			if err != nil {
				log.Println(err)
				continue
			}

			var division, homeTeam string

			sj := htmlquery.FindOne(item, `//div[@class="subject-group"]`)

			if sj != nil {
				division = htmlquery.InnerText(sj)
				homeTeam, err = parser.QueryInnerText(item, `//div[contains(@class,"subject-owner")]`)
				if err != nil {
					log.Fatal("subject owner error ", err, content)
				}
			} else {
				d, err := parser.QueryInnerText(item, `//div[contains(@class,"subject-owner")]`)
				if err != nil {
					log.Fatal("subject owner error ", err, content)
				}
				// re := regexp.MustCompile(`^(U\w+ [A-Z]{1,}) (.+)$`)
				// rs := re.FindStringSubmatch(d)
				//
				// if rs == nil {
				// 	log.Fatal("failed to parse team: ", d)
				// }
				// division = rs[1]
				// homeTeam = rs[2]
				division, homeTeam = d, d
			}

			subjectText, err := htmlquery.Query(item, `//div[contains(@class, "subject-text")]`)

			if err != nil {
				log.Println(err)
				continue
			}

			ch := subjectText.FirstChild
			guestTeam := strings.Replace(htmlquery.InnerText(ch), "@ ", "", -1)

			if !homeGame {
				homeTeam, guestTeam = guestTeam, homeTeam
			}

			location, err := parser.QueryInnerText(item, `//div[contains(@class,"location")]`)

			item = htmlquery.Find(parent, `div[1]//a[@class="remote" or @class="local"]`)[0]
			var url string
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
				if url[:4] != "http" {
					url = baseURL + url
				}
				go func(url string, location string, wg *sync.WaitGroup, lock *sync.Mutex) {
					defer wg.Done()
					address := parser.GetVenueAddress(url, class)
					address = strings.Replace(address, location, "", 1)

					lock.Lock()
					result = append(result, []string{ymd + " " + timeval, Site, homeTeam, guestTeam, location, division, address})
					lock.Unlock()
				}(url, location, wg, lock)
			} else {
				result = append(result, []string{ymd + " " + timeval, Site, homeTeam, guestTeam, location, division, ""})
			}
		}
	}
	wg.Wait()
	return result
}
