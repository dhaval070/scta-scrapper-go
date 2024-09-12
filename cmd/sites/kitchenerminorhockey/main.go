package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"flag"

	"calendar-scrapper/config"
	"calendar-scrapper/dao/model"
	"calendar-scrapper/pkg/parser"
	"calendar-scrapper/pkg/parser1"
	"calendar-scrapper/pkg/repository"
	"calendar-scrapper/pkg/writer"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

const SITE = "kitchenerminorhockey"
const HOME_TEAM = "kitchenerminorhockey"
const BASE_URL = "https://kitchenerminorhockey.com/"

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

	var result = parser1.ParseSchedules(doc, SITE, BASE_URL, HOME_TEAM)

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

func gameDetails(url string) string {
	resp, err := parser.Client.Get(url)
	if err != nil {
		log.Println(err)
		return ""
	}
	defer resp.Body.Close()

	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		log.Println("error getting "+url, err)
		return ""
	}

	// re := regexp.MustCompile(`/Teams/`)

	var url1 string
	nodes := htmlquery.Find(doc, `//div[contains(@class,"game-details-tabs")]//a`)

	for _, n := range nodes {
		url1 = parseHref(n)

		if htmlquery.InnerText(n) == "More Venue Details" {
			// if re.Match([]byte(url1)) {
			return venueDetails(url1)
		}
	}
	return ""
}

func venueDetails(url string) (address string) {
	resp, err := parser.Client.Get(BASE_URL + url)
	if err != nil {
		log.Println("error get "+url, err)
		return ""
	}
	defer resp.Body.Close()

	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		log.Println("error parsing "+url, err)
		return ""
	}

	node := htmlquery.FindOne(doc, `//div[@class="month"]/following-sibling::div/div/div`)
	if node == nil {
		log.Println("error venue detail not found ", BASE_URL+url)
		return ""
	}
	address = htmlquery.InnerText(node)
	return address
}

func parseHref(node *html.Node) string {
	if node == nil {
		return ""
	}
	for _, attr := range node.Attr {
		if attr.Key == "href" {
			return attr.Val
		}
	}
	return ""
}
