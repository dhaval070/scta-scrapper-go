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

const SITE = "bghc"
const HOME_TEAM = "bghc"
const BASE_URL = "https://bghc.ca/"

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

	var result = parseSchedules(doc, SITE, BASE_URL, HOME_TEAM)

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

func parseSchedules(doc *html.Node, Site, baseURL, homeTeam string) [][]string {
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

			var gameType = htmlquery.InnerText(htmlquery.FindOne(item, `div[2]`))

			if gameType == "Tournament" || gameType == "Hosted Tournament" {
				// parser1.ParseTournament(parent)
				continue
			}
			//Practice or SharedPractice
			if strings.Contains(gameType, "Practice") {
				continue
			}

			if strings.Contains(gameType, "Away") {
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

			// log.Println(gameType)
			item = htmlquery.FindOne(parent, `//div[1]/div[2]/a`)
			url := parseHref(item)

			if url != "" {
				wg.Add(1)
				if url[:4] != "http" {
					url = baseURL + url
				}

				go func(url string, location string, wg *sync.WaitGroup, lock *sync.Mutex) {
					defer wg.Done()
					address := gameDetails(url)
					if address == "" {
						log.Println("addr not found ", url)
					} else {
						log.Println(url, address)
					}
					lock.Lock()
					result = append(result, []string{ymd + " " + timeval, Site, hm, guestTeam, location, division, address})
					lock.Unlock()
				}(url, location, wg, lock)
			} else {
				log.Println("url not found")
				result = append(result, []string{ymd + " " + timeval, Site, hm, guestTeam, location, division, ""})
			}
		}
	}
	wg.Wait()
	return result
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
