package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
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

const SITE = "sdhockey"
const HOME_TEAM = "sdahockey"

func parseGroups(doc *html.Node) map[string]string {
	links := htmlquery.Find(doc, `//div[@class="site-list"]//a`)

	var groups = make(map[string]string)

	for _, n := range links {
		href := parser.GetAttr(n, "href")
		division := htmlquery.InnerText(n)

		re := regexp.MustCompile(`Teams/(.+)/`)

		parts := re.FindAllStringSubmatch(href, -1)
		if parts == nil {
			log.Fatal("failed to parse group link", href)
		}
		groups[division] = parts[0][1]
	}
	return groups
}

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

	doc, err = htmlquery.LoadURL(fmt.Sprintf("https://sdhockey.ca/Calendar/?Month=%d&Year=%d", mm, yyyy))
	if err != nil {
		panic(err)
	}

	var result = parseSchedules(doc)

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

func parseSchedules(doc *html.Node) [][]string {
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

			if !strings.Contains(content, "Home Game") && !strings.Contains(content, "Away Game") {
				continue
			}

			timeval, err := parser.ParseTime(content)
			if err != nil {
				log.Fatal(err)
				continue
			}

			division, err := parser.QueryInnerText(item, `div[3]/div[1]`)
			if err != nil {
				log.Println(err)
				continue
			}

			guestStr, err := parser.QueryInnerText(item, `//div[contains(@class, "subject-text")]`)
			if err != nil {
				log.Fatal(err)
			}

			guestTeam := strings.Replace(guestStr, "@ ", "", -1)

			location, err := parser.QueryInnerText(item, `//div[contains(@class,"location")]`)

			item = htmlquery.Find(parent, `div[1]//a[@class="remote" or @class="local"]`)[0]
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
					url = "https://sdhockey.ca" + url
				}
				go func(url string, location string, wg *sync.WaitGroup, lock *sync.Mutex) {
					defer wg.Done()
					address := parser.GetVenueAddress(url, class)
					address = strings.Replace(address, location, "", 1)

					lock.Lock()
					result = append(result, []string{ymd + " " + timeval, SITE, HOME_TEAM, guestTeam, location, division, address})
					lock.Unlock()
				}(url, location, wg, lock)
			} else {
				result = append(result, []string{ymd + " " + timeval, SITE, HOME_TEAM, guestTeam, location, division, ""})
			}
		}
	}
	wg.Wait()
	return result
}
func parseMonthYear(dt string) (int, int) {
	re := regexp.MustCompile(`^[0-9]{6}$`)

	if !re.Match([]byte(dt)) {
		panic("invalid format mmyyyy input")
	}

	mm, err := strconv.Atoi(dt[:2])
	if err != nil {
		panic(err)
	}

	yyyy, err := strconv.Atoi(dt[2:])
	if err != nil {
		panic(err)
	}
	return mm, yyyy
}
