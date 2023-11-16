package main

import (
	"calendar-scrapper/config"
	"calendar-scrapper/dao/model"
	"calendar-scrapper/pkg/parser"
	"calendar-scrapper/pkg/repository"
	"calendar-scrapper/pkg/writer"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"flag"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

const SITE = "alliancehockey"

func main() {
	infile := flag.String("infile", "", "local html filename")
	date := flag.String("date", "", "calendar month and year in format: mmyyyy")
	outfile := flag.String("outfile", "", "output filename")
	importLocations := flag.Bool("import-locations", false, "import site locations")

	flag.Parse()

	var doc *html.Node
	var err error
	var mm, yyyy int

	today := time.Now()
	mm = int(today.Month())
	yyyy = int(today.Year())

	if *infile != "" {
		doc, err = htmlquery.LoadDoc(*infile)
	} else {
		if *date != "" {
			mm, yyyy = parseMonthYear(*date)
		}

		url := fmt.Sprintf("https://alliancehockey.com/Schedule/?Month=%d&Year=%d", mm, yyyy)
		log.Println(url)
		doc, err = htmlquery.LoadURL(url)
	}

	if err != nil {
		log.Fatal(err)
	}

	result := ParseSchedules(doc, mm, yyyy)

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
		repo.ImportLoc(locations)
	}

	sort.Sort(parser.ByDate(result))

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

func ParseSchedules(doc *html.Node, mm, yyyy int) [][]string {
	nodes := htmlquery.Find(doc, `//div[contains(@class, "day-details")]`)
	if len(nodes) == 0 {
		log.Println(htmlquery.OutputHTML(doc, true))
		panic("day details not found")
	}

	node := nodes[0]
	var result = [][]string{}

	//id := GetAttr(node, "id")
	//dt, ymd := ParseId(id)
	listItems := htmlquery.Find(node, `//div[contains(@class, "event-list-item")]/div`) // `div[2]`)

	// infoItems := htmlquery.Find(node, `//div[contains(@class, "event-list-item")]/div/div[1]`)

	dateMatch := regexp.MustCompile(`[0-9]{2}$`)

	var lock = &sync.Mutex{}
	var wg = &sync.WaitGroup{}

	for _, parent := range listItems {
		items := htmlquery.Find(parent, `div[2]`)
		item := items[0]
		content := htmlquery.OutputHTML(item, true)

		if strings.Contains(content, "All Day") || strings.Contains(content, "time-secondary") || strings.Contains(content, "Cancelled") {
			log.Println("skipping")
			continue
		}
		timeval := parser.ParseTime(content)
		sDate, err := parser.QueryInnerText(item, `//div[@class="day_of_month"]`)
		if err != nil {
			log.Println(err)
			continue
		}

		dayOfMonth := dateMatch.FindString(sDate)
		if dayOfMonth == "" {
			log.Println("day of month not found")
			continue
		}
		ymd := fmt.Sprintf("%d-%d-%s", yyyy, mm, dayOfMonth)

		division, err := parser.QueryInnerText(item, `//div[@class="subject-group"]`)
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
				address = getVenueAddress(url)
				lock.Lock()
				result = append(result, []string{ymd + " " + timeval, SITE, homeTeam, guestTeam, location, division, address})
				lock.Unlock()
			}(url, wg, lock)
		} else {
			result = append(result, []string{ymd + " " + timeval, SITE, homeTeam, guestTeam, location, division, address})
		}
	}
	wg.Wait()
	return result
}

func getVenueAddress(url string) string {
	doc, err := htmlquery.LoadURL(url)
	if err != nil {
		log.Println("error getting "+url, err)
		return ""
	}

	item := htmlquery.FindOne(doc, `//div[@class="container"]/div/div/h2/small[2]`)
	if item == nil {
		log.Println("address node not found")
		return ""
	}
	address := htmlquery.InnerText(item)
	log.Println(url + ":" + address)
	return address
}
