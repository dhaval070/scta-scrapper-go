package parser1

import (
	"calendar-scrapper/pkg/parser"
	"log"
	"strings"
	"sync"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

func ParseSchedules(doc *html.Node, Site, baseURL, homeTeam string) [][]string {
	nodes := htmlquery.Find(doc, `//div[contains(@class, "day-details")]`)

	var result = [][]string{}

	var lock = &sync.Mutex{}
	var wg = &sync.WaitGroup{}

	for _, node := range nodes {
		listItems := htmlquery.Find(node, `//div[contains(@class, "event-list-item")]/div`)

		var id string

		for _, v := range node.Attr {
			if v.Key == "id" {
				id = v.Val
				break
			}
		}
		if id == "" {
			log.Fatal("id not found")
		}
		ymd := parser.ParseId(id)

		for _, parent := range listItems {
			item := htmlquery.FindOne(parent, `div[2]`)
			content := htmlquery.OutputHTML(item, true)

			var homeGame bool

			if strings.Contains(strings.ToUpper(content), "HOME GAME") {
				homeGame = true
			} else if !strings.Contains(strings.ToUpper(content), "AWAY GAME") {
				// neither home game or away game then skip
				continue
			}

			timeval, err := parser.ParseTime(content)
			if err != nil {
				log.Fatal(err)
				continue
			}

			division, err := parser.QueryInnerText(item, `div[3]/div[1]`)
			if err != nil {
				log.Println(err)
				continue
			}

			guestStr, err := parser.QueryInnerText(item, `//div[contains(@class, "subject-text")]`)
			if err != nil {
				log.Fatal(err)
			}

			guestTeam := strings.Replace(guestStr, "@ ", "", -1)

			hm := homeTeam

			if !homeGame {
				hm, guestTeam = guestTeam, homeTeam
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
					result = append(result, []string{ymd + " " + timeval, Site, hm, guestTeam, location, division, address})
					lock.Unlock()
				}(url, location, wg, lock)
			} else {
				result = append(result, []string{ymd + " " + timeval, Site, hm, guestTeam, location, division, ""})
			}
		}
	}
	wg.Wait()
	return result
}
