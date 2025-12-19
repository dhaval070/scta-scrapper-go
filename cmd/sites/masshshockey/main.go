package main

import (
	"calendar-scrapper/pkg/cmdutil"
	"calendar-scrapper/pkg/writer"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var cl = http.DefaultClient

const baseUrl = "https://hnib-service-ocx2eq7snq-uk.a.run.app/api/v1/client/10/games?seasonId=32885&gender=%s&size=200&sort=date&sort=time&page=%d"

const SITE = "masshshockey"

type Response struct {
	Content    []Game `json:"content"`
	TotalPages int    `json:"totalPages"`
	Last       bool   `json:"last"`
}

type Game struct {
	Id       int64  `json:"id"`
	Date     string `json:"date"`
	Time     string `json:"time"`
	Location string `json:"location"`
	HomeTeam string `json:"homeTeam"`
	AwayTeam string `json:"awayTeam"`
}

func main() {
	flags := cmdutil.ParseCommonFlags()
	flag.Parse()

	result := getSchedles("men")
	if len(result) == 0 {
		log.Println("no games found")
		return
	}
	result = append(result, getSchedles("women")...)

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

func getSchedles(gender string) [][]string {
	page := 0
	var result = [][]string{}
	for {
		url := fmt.Sprintf(baseUrl, gender, page)

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			log.Println("failed create http request")
			return nil
		}
		req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/112.0")
		req.Header.Add("Accept", "application/json")

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
		var res Response

		if err = json.Unmarshal(data, &res); err != nil {
			log.Println("error in json decode %w", err)
			return nil
		}

		log.Println("count", len(res.Content))

		if len(res.Content) == 0 {
			log.Println("content array empty")
			log.Println(string(data))
			break
		}
		for _, g := range res.Content {
			dt := g.Date
			dt = dt + " " + g.Time[0:5]
			result = append(result, []string{
				dt, SITE, g.HomeTeam, g.AwayTeam, g.Location, "", "",
			})
		}
		if res.Last {
			log.Println("last page")
			break
		}
		page += 1

	}
	return result
}
