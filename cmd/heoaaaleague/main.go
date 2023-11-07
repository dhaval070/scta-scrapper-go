package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"flag"

	"calendar-scrapper/config"
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

	mm := (*today)[0:4]
	yyyy := (*today)[4:]
	url := fmt.Sprintf("https://heoaaaleague.ca/Schedule/?Month=%s&Year=%s", mm, yyyy)

	doc, err = htmlquery.LoadURL(url)
	if err != nil {
		log.Fatal("load calendar url", err)
	}

	result := parseSchedules(doc, *today)

	if *importLocations {
		config.Init("config", ".")
		cfg := config.MustReadConfig()

		var locations []string
		for _, r := range result {
			locations = append(locations, r[4])
		}

		repo := repository.NewRepository(cfg).Site(SITE)
		repo.ImportLocations(locations)
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

	for _, node := range nodes {
		listItems := htmlquery.Find(node, `//div[contains(@class, "event-list-item")]/div/div[2]`)
		for _, item := range listItems {
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
			fmt.Sscanf("\w+ %d")
			dom := txt[-2:]

			ymd := fmt.Sprintf("%s-%s-%s", today[:4], today[4:]) //, timeval)

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
			result = append(result, []string{ymd + " " + timeval, "", homeTeam, guestTeam, location, division})
		}
	}
	return result
}
