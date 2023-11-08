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

const SITE = "heoaaaleague"

func main() {
	ymd := time.Now().Format("20060102")

	today := flag.String("today", ymd, "parse from date(yyyymmdd)")
	outfile := flag.String("outfile", "", "output filename")
	importLocations := flag.Bool("import-locations", false, "import site locations")

	flag.Parse()
	log.Println(*today)

	var doc *html.Node
	var err error

	mm := (*today)[4:6]
	yyyy := (*today)[:4]
	url := fmt.Sprintf("https://heoaaaleague.ca/Schedule/?Month=%s&Year=%s", mm, yyyy)

	log.Println(url)

	doc, err = htmlquery.LoadURL(url)
	if err != nil {
		log.Fatal("load calendar url", err)
	}

	result := parseSchedules(doc, *today)

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

func parseSchedules(doc *html.Node, today string) [][]string {
	nodes := htmlquery.Find(doc, `//div[contains(@class, "day-details")]`)

	var result = [][]string{}

	var lock = &sync.Mutex{}
	var wg = &sync.WaitGroup{}

	for _, node := range nodes {
		listItems := htmlquery.Find(node, `//div[contains(@class, "event-list-item")]/div`)
		for _, parent := range listItems {
			items := htmlquery.Find(parent, `div[2]`)
			item := items[0]
			content := htmlquery.OutputHTML(item, true)

			if strings.Contains(content, "All Day") || strings.Contains(content, "time-secondary") || strings.Contains(content, "Cancelled") {
				continue
			}

			timeval := parser.ParseTime(content)

			txt, err := parser.QueryInnerText(item, `//div[@class="day_of_month"]`)
			if err != nil {
				log.Println(err)
				continue
			}
			// var dom string
			dom := txt[4:]

			ymd := fmt.Sprintf("%s-%s-%s", today[:4], today[4:], dom) //, timeval)

			division, err := parser.QueryInnerText(item, `//span[@class="game_no"]`)
			if err != nil {
				log.Println(err)
				continue
			}
			guestTeam, err := parser.QueryInnerText(item, `//div[contains(@class, "subject-owner")]`)
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
			location, err := parser.QueryInnerText(item, `//div[@class="location remote"]`)

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
					address = parser.GetVenueAddress(url)
					lock.Lock()
					result = append(result, []string{ymd + " " + timeval, "", homeTeam, guestTeam, location, division, address})
					lock.Unlock()
				}(url, wg, lock)
			} else {
				result = append(result, []string{ymd + " " + timeval, "", homeTeam, guestTeam, location, division, address})
			}
		}
	}
	wg.Wait()
	return result
}
