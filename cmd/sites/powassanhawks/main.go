package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
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

const SITE = "powassanhawks"
const BASE_URL = "https://powassanhawks.com/"

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

	doc, err = htmlquery.LoadURL(fmt.Sprintf(BASE_URL+"Calendar/?Month=%d&Year=%d", mm, yyyy))
	if err != nil {
		panic(err)
	}

	var result = parseSchedules(doc, SITE, BASE_URL)

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

func parseSchedules(doc *html.Node, Site, baseURL string) [][]string {
	nodes := htmlquery.Find(doc, `//div[contains(@class, "day-details")]`)

	var result = [][]string{}

	var lock = &sync.Mutex{}
	var wg = &sync.WaitGroup{}

	// https://powassanhawks.ca has three types of games home, local and away.
	reHome := regexp.MustCompile("(?i)(home|local) game")
	reAway := regexp.MustCompile("(?i)away game")

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

			if reHome.MatchString(content) {
				homeGame = true
			} else if !reAway.MatchString(content) {
				// neither home game or away game then skip
				continue
			}

			timeval, err := parser.ParseTime(htmlquery.InnerText(htmlquery.FindOne(item, `div/div[@class="time-primary"]`)))

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

			item = htmlquery.FindOne(parent, `div[1]//a[@class="remote" or @class="local"]`)
			if item == nil {
				log.Fatal("failed to get venue link", htmlquery.OutputHTML(parent, true))
			}
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
