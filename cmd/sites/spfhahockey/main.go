package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"flag"

	"calendar-scrapper/config"
	"calendar-scrapper/dao/model"
	"calendar-scrapper/pkg/parser"
	"calendar-scrapper/pkg/repository"
	"calendar-scrapper/pkg/writer"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

const SITE = "spfhahockey"
const BASE_URL = "https://spfhahockey.com/Schedule/"
const HOME_TEAM = "sun parlour"

func main() {
	date := flag.String("date", "", "calendar month and year in format: mmyyyy")
	outfile := flag.String("outfile", "", "output filename")
	importLocations := flag.Bool("import-locations", false, "import site locations")

	flag.Parse()

	var doc *html.Node
	var err error
	var mm, yyyy int

	if *date == "" {
		today := time.Now()
		mm = int(today.Month())
		yyyy = int(today.Year())

	} else {
		mm, yyyy = parser.ParseMonthYear(*date)
	}

	doc, err = htmlquery.LoadURL(fmt.Sprintf(BASE_URL+"?Month=%d&Year=%d", mm, yyyy))
	if err != nil {
		panic(err)
	}

	var result = parseSchedules(doc, mm, yyyy)

	if *importLocations {
		config.Init("config", ".")
		cfg := config.MustReadConfig()

		var locations = make([]model.SitesLocation, 0, len(result))
		for _, r := range result {
			log.Printf("%+v\n", r)

			l := model.SitesLocation{
				Location: r[4],
				Address:  r[6],
			}
			locations = append(locations, l)
		}

		repo := repository.NewRepository(cfg).Site(SITE)
		if err = repo.ImportLoc(locations); err != nil {
			log.Fatal(err)
		}
	}
	if *outfile != "" {
		fh, err := os.Create(*outfile)
		if err != nil {
			log.Fatal(err)
		}
		writer.WriteCsv(fh, result)
	} else {
		log.Println(result)
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
					url = "https://spfhahockey.com/" + url
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
