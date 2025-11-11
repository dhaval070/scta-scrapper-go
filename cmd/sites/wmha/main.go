package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"calendar-scrapper/pkg/parser"
	"calendar-scrapper/pkg/cmdutil"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

const SITE = "wmha"
const BASE_URL = "https://wmha.net/Schedule/"
const HOME_TEAM = "Windsor Spitfires"

func main() {
	flags := cmdutil.ParseCommonFlags()

	var doc *html.Node
	var err error
	var mm, yyyy int

	if *flags.Date == "" {
		today := time.Now()
		mm = int(today.Month())
		yyyy = int(today.Year())

	} else {
		mm, yyyy = parser.ParseMonthYear(*flags.Date)
	}

	doc, err = htmlquery.LoadURL(fmt.Sprintf(BASE_URL+"?Month=%d&Year=%d", mm, yyyy))
	if err != nil {
		panic(err)
	}

	var result = parseSchedules(doc, mm, yyyy)

	if *flags.ImportLocations {
		if err := cmdutil.ImportLocations(SITE, result); err != nil {
			log.Fatal(err)
		}
	}
	if err := cmdutil.WriteOutput(*flags.Outfile, result); err != nil {
		log.Fatal(err)
	}
}

func parseSchedules(doc *html.Node, mm, yyyy int) [][]string {
	nodes := htmlquery.Find(doc, `//div[contains(@class, "day-details")]`)

	var result = [][]string{}

	var lock = &sync.Mutex{}
	var wg = &sync.WaitGroup{}

	for _, node := range nodes {
		listItems := htmlquery.Find(node, `//div[contains(@class, "event-list-item")]/div`)

		for _, parent := range listItems {
			item := htmlquery.FindOne(parent, `div[2]`)
			content := htmlquery.OutputHTML(item, true)

			if strings.Contains(content, "Tournament") {
				log.Println("skipping tournament")
				continue
			}

			timeval, err := parser.ParseTime(content)
			if err != nil {
				log.Println(err)
				continue
			}

			txt, err := parser.QueryInnerText(item, `//div[@class="day_of_month"]`)
			if err != nil {
				log.Println(err)
				continue
			}
			// var dom string
			dom := txt[4:]

			ymd := fmt.Sprintf("%d-%d-%s", yyyy, mm, dom) //, timeval)
			var division, homeTeam, guestTeam string

			division, err = parser.QueryInnerText(item, `//div[contains(@class,"subject-owner")]`)
			if err != nil {
				log.Fatal("subject owner error ", err, content)
			}

			subjectText, err := htmlquery.Query(item, `//div[contains(@class, "subject-text")]`)

			if err != nil {
				log.Println(err)
				continue
			}

			ch := htmlquery.InnerText(subjectText.FirstChild)
			if ch[0] == '@' {
				guestTeam = HOME_TEAM
				homeTeam = ch[2:]
			} else if ch[0:3] == "vs " {
				homeTeam = HOME_TEAM
				guestTeam = ch[3:]
			} else {
				log.Fatal("failed to parse teams")
			}

			location, err := parser.QueryInnerText(item, `//div[contains(@class,"location")]`)

			item = htmlquery.Find(parent, `div[1]//a[@class="remote" or @class="local"]`)[0]
			var url string
			var class string
			var address string

			for _, attr := range item.Attr {
				if attr.Key == "href" {
					url = attr.Val
					break
				} else if attr.Key == "class" {
					class = attr.Val
				}
			}

			if url != "" {
				if url[0:4] != "http" {
					url = "https://wmha.net/" + url
				}
				wg.Add(1)
				go func(url string, wg *sync.WaitGroup, lock *sync.Mutex) {
					defer wg.Done()
					address = parser.GetVenueAddress(url, class)
					lock.Lock()
					result = append(result, []string{ymd + " " + timeval, SITE, homeTeam, guestTeam, location, division, address})
					lock.Unlock()
				}(url, wg, lock)
			} else {
				result = append(result, []string{ymd + " " + timeval, SITE, homeTeam, guestTeam, location, division, address})
			}
		}
	}
	wg.Wait()
	return result
}
