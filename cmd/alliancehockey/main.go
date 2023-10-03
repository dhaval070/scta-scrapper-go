package main

import (
	"calendar-scrapper/config"
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

		var locations []string
		for _, r := range result {
			locations = append(locations, r[4])
		}

		repo := repository.NewRepository(cfg).Site(SITE)
		repo.ImportLocations(locations)
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
	listItems := htmlquery.Find(node, `//div[contains(@class, "event-list-item")]/div/div[2]`)

	dateMatch := regexp.MustCompile(`[0-9]{2}$`)

	for _, item := range listItems {
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

		result = append(result, []string{ymd + " " + timeval, "", homeTeam, guestTeam, location, division})
	}
	return result
}
