package main

import (
	"fmt"
	"log"
	"time"

	"calendar-scrapper/pkg/parser"
	"calendar-scrapper/pkg/cmdutil"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

const SITE = "londonjuniorknights"
const HOME_TEAM = "londonjuniorknights"
const BASE_URL = "https://londonjuniorknights.com/"

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

	doc, err = htmlquery.LoadURL(fmt.Sprintf(BASE_URL+"Calendar/?Month=%d&Year=%d", mm, yyyy))
	if err != nil {
		panic(err)
	}

	var result = parseSchedules(doc, SITE, BASE_URL, HOME_TEAM)

	if *flags.ImportLocations {
		if err := cmdutil.ImportLocations(SITE, result); err != nil {
			log.Fatal(err)
		}
	}
	if err := cmdutil.WriteOutput(*flags.Outfile, result); err != nil {
		log.Fatal(err)
	}
}

func parseSchedules(doc *html.Node, Site, baseURL, homeTeam string) [][]string {
	cfg := parser.DayDetailsConfig{
		TournamentCheckExact: false,
		LogErrors:            true,
		GameDetailsFunc:      gameDetails,
	}
	return parser.ParseDayDetailsSchedule(doc, Site, baseURL, homeTeam, cfg)
}

func gameDetails(url string) string {
	return parser.GetGameDetailsAddress(url, BASE_URL)
}
