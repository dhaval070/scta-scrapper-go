package main

import (
	"bytes"
	"calendar-scrapper/pkg/cmdutil"
	"calendar-scrapper/pkg/htmlutil"
	"calendar-scrapper/pkg/writer"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

var cl = http.DefaultClient

const baseUrl = "https://neutralzone.com/prep-boys/schedule/"
const SITE = "neutralzone"

func main() {
	flags := cmdutil.ParseCommonFlags()
	flag.Parse()

	dates := []string{
		"2026-01-05",
		"2025-12-29",
		"2025-12-15",
	}

	// wg := sync.WaitGroup{}
	// wg.Add(len(dates))

	result := [][]string{}
	for _, dt := range dates {
		res := getSchedules(dt)
		result = append(result, res...)
	}

	if *flags.ImportLocations {
		if err := cmdutil.ImportLocations(SITE, result); err != nil {
			log.Fatal(err)
		}
	}

	if *flags.Outfile != "" {
		var fh *os.File
		var err error

		if *flags.Outfile == "-" {
			fh = os.Stdout
		} else {
			fh, err = os.Create(*flags.Outfile)
			if err != nil {
				log.Fatal(err)
			}
			defer fh.Close()
		}
		if err = writer.WriteCsv(fh, result); err != nil {
			panic(err)
		}
	}
}

func getSchedules(date string) [][]string {
	body := strings.NewReader("week=" + date)

	req, err := http.NewRequest(http.MethodPost, baseUrl, body)
	if err != nil {
		log.Println("error in creating http request")
		return nil
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/112.0")
	req.Header.Add("Accept", "text/html")

	resp, err := cl.Do(req)

	if err != nil {
		log.Println("http request failed %w", err)
		return nil
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Println("failed to read body %w", err)
		return nil
	}

	doc, err := htmlquery.Parse(bytes.NewReader(data))
	if err != nil {
		log.Println("failed to parse document %w", err)
		return nil
	}
	var result = [][]string{}

	for _, entry := range htmlquery.Find(doc, `//table[@id="table-schedule"]/tbody/tr`) {
		dt := htmlquery.FindOne(entry, `td[@class="date"]`)
		if dt == nil {
			log.Println("date not found")
			continue
		}

		ymd := htmlutil.GetAttr(dt, "data-sort")
		if ymd == "" {
			log.Println("ymd not found")
			continue
		}

		homeTeam := parseTeam(entry, "home-team")
		if homeTeam == "" {
			log.Println("home team empty")
			continue
		}
		awayTeam := parseTeam(entry, "away-team")
		if awayTeam == "" {
			log.Println("away team empty")
			continue
		}

		loc := htmlquery.FindOne(entry, `td[@class="location"]`)
		if loc == nil {
			log.Println("td.location not found")
			continue
		}
		location := htmlquery.InnerText(loc)

		tt := htmlquery.FindOne(entry, `td[@class="time"]`)
		if tt == nil {
			log.Println("td.time not found")
			continue
		}
		tt1 := htmlquery.InnerText(tt)
		if tt1 == "" {
			continue
		}
		tt2, err := time.Parse(`2006-01-02 3:04 PM`, ymd+" "+tt1)
		if err != nil {
			log.Println("failed to parse time - ", ymd, tt1)
			continue
		}
		// division and address are not available
		result = append(result, []string{
			tt2.Format("2006-01-02 15:04"),
			SITE,
			homeTeam,
			awayTeam,
			location,
			"", "",
		})
	}
	return result
}

func parseTeam(node *html.Node, teamType string) string {
	teamlink := htmlquery.FindOne(node, fmt.Sprintf(`td[@class="%s"]/a[@class="team-link"]`, teamType))
	if teamlink == nil {
		return ""
	}

	return htmlquery.InnerText(teamlink)
}
