package parser1

import (
	"calendar-scrapper/pkg/htmlutil"
	"calendar-scrapper/pkg/parser"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

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

			var homeGame = true

			gameType := htmlquery.InnerText(htmlquery.FindOne(item, `div[2]/div/div`))
			if gameType == "Event" ||
				gameType == "tryOut" ||
				strings.Contains(content, "Practice") ||
				strings.Contains(gameType, "Tournament") {
				continue
			}

			if strings.Contains(strings.ToUpper(content), "AWAY GAME") ||
				strings.Contains(strings.ToUpper(content), "AWAY TOURNAMENT") ||
				strings.Contains(strings.ToUpper(content), "AWAY EXHIBITION") {
				homeGame = false
			}

			timeval, err := parser.ParseTime(content)
			if err != nil {
				log.Println(err)
				continue
			}

			division, err := parser.QueryInnerText(item, `div[3]/div[1]`)
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
			guestTeam := strings.Replace(htmlquery.InnerText(ch), "@ ", "", 1)
			guestTeam = strings.Replace(guestTeam, "vs ", "", 1)

			hm := homeTeam

			if !homeGame {
				hm, guestTeam = guestTeam, homeTeam
			}

			location, err := parser.QueryInnerText(item, `//div[contains(@class,"location")]`)
			if err != nil {
				log.Printf("error %v\n", err)
				continue
			}

			item = htmlquery.FindOne(parent, `div[1]//a[@class="remote" or @class="local"]`)
			var url string
			var class string

			if item == nil {
				log.Println("error parsing venue url")
				continue
			}

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
					address, err := parser.VenueFetcher.Fetch(url, class)
					if err != nil {
						log.Println("Error fetching venue address:", err)
						address = ""
					}
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

func ParseTournament(doc *html.Node, tournamentId string) [][]string {
	// node := htmlquery.Find(doc, fmt.Sprintf(`//label[@for="team_tournament_%s"]/following-sibling::div/div[contains(@class, "tournament-details")]/div[@class="tabs-content"]`, tournamentId))

	details := htmlquery.FindOne(doc, fmt.Sprintf(`//label[@for="team_tournament_%s"]/following-sibling::div/div[contains(@class, "tournament-details")]`, tournamentId))

	nodes := htmlquery.Find(details, `ul/li/div[@class="accordion-content"]/div/div`)

	result := [][]string{}

	for _, parent := range nodes {
		sdate := htmlquery.InnerText(htmlquery.FindOne(parent, `h4`))
		log.Println(sdate)

		for _, row := range htmlquery.Find(parent, `//div[contains(@class,"event-list-item")]`) {
			tt, err := parser.ParseTime(htmlquery.InnerText(htmlquery.FindOne(row, `//div[@class="time-primary"]`)))

			if err != nil {
				log.Println("error parsing time")
				continue
			}

			dt, err := time.Parse("Mon, Jan 2 2006 15:04", fmt.Sprintf(sdate+" %d "+tt, time.Now().Year()))

			if err != nil {
				log.Println(err)
				continue
			}

			guestTeam := htmlquery.InnerText(htmlquery.FindOne(parent, `//div[@class="subject-text"]`).FirstChild)
			guestTeam = strings.Replace(guestTeam, "vs ", "", 1)
			guestTeam = strings.Replace(guestTeam, "@ ", "", 1)

			loc := htmlquery.InnerText(htmlquery.FindOne(parent, `//div[contains(@class, "location")]`))

			link := htmlquery.FindOne(row, `div/div/div/a`)
			var venueUrl string

			if link != nil {
				venueUrl = htmlutil.GetAttr(link, "href")

			}
			result = append(result, []string{dt.Format("2006-01-02 15:04"), guestTeam, loc, venueUrl})

		}
	}
	return result
}
