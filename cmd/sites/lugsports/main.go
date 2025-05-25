package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"calendar-scrapper/config"
	"calendar-scrapper/dao/model"
	"calendar-scrapper/internal/webdriver"
	"calendar-scrapper/pkg/repository"

	"github.com/antchfx/htmlquery"
	"github.com/tebeka/selenium"
	"golang.org/x/net/html"
)

var siteUrl = "https://www.lugsports.com/stats#/1709/schedule?season_id=8683&all"

const SITE = "lugsports"

func main() {
	infile := flag.String("infile", "", "local html filename")
	importLocations := flag.Bool("import-locations", false, "import site locations")
	outfile := flag.String("outfile", "", "output filename")
	// date flag is not used but have to add here to make this command compatible with other sites and use with run.sh and run-all.sh
	_ = flag.String("date", "", "calendar month and year in format: mmyyyy")

	flag.Parse()

	var source string
	var doc *html.Node
	var err error

	if *infile != "" {
		doc, err = htmlquery.LoadDoc(*infile)
		if err != nil {
			panic("failed to load file")
		}
	} else {
		service, err := selenium.NewChromeDriverService("./chromedriver", 4444)
		if err != nil {
			log.Fatal("Error:", err)
		}
		defer service.Stop()

		driver := webdriver.GetWebDriver()

		err = driver.Get(siteUrl)
		if err != nil {
			log.Println("failed to get %s: %w", siteUrl, err)
		}

		time.Sleep(8 * time.Second)
		source, err = driver.PageSource()

		if err != nil {
			log.Println("pageSource error %s: %w", siteUrl, err)
		}

		doc, err = htmlquery.Parse(strings.NewReader(source))
		if err != nil {
			log.Println(source)
			panic("failed to parse html")
		}
	}

	node := htmlquery.FindOne(doc, `//partial[@slug="stats/schedule/table"]/div//table[@class="schedule"]/tbody`)
	if node == nil {
		panic("schedule not found")
	}

	var content string
	for _, attr := range node.Attr {
		if attr.Key == "ng-init" {
			content = attr.Val
			break
		}
	}
	if content == "" {
		panic("ng-init not found")
	}
	content = strings.Replace(content, "ctrl.schedule=", "", 1)

	var result []Event

	if err = json.Unmarshal([]byte(content), &result); err != nil {
		panic(fmt.Errorf("failed to unmarshal schedule json %w", err))
	}
	for i, v := range result {
		address, err := url.QueryUnescape(v.FacilityAddress)
		if err != nil {
			panic(fmt.Errorf("failed to unescape address %s ", v.FacilityAddress))
		}
		result[i].FacilityAddress = address
	}

	if *importLocations {
		config.Init("config", ".")
		cfg := config.MustReadConfig()

		var locations = make([]model.SitesLocation, 0, len(result))
		for _, r := range result {
			l := model.SitesLocation{
				Location: r.Facility,
				Address:  r.FacilityAddress,
			}
			locations = append(locations, l)
		}

		repo := repository.NewRepository(cfg).Site(SITE)
		repo.ImportLoc(locations)
	}

	if *outfile != "" {
		fh, err := os.Create(*outfile)
		if err != nil {
			log.Fatal(err)
		}
		WriteEvents(fh, result)
		fh.Close()
	}
}

func WriteEvents(w io.Writer, data []Event) error {
	var err error
	ww := csv.NewWriter(w)
	for _, rec := range data {
		row := []string{
			rec.Datetime.Format("2006-1-2 15:04"),
			SITE,
			rec.HomeTeam,
			rec.AwayTeam,
			rec.Facility,
			rec.HomeDivision,
			rec.FacilityAddress,
		}
		if err = ww.Write(row); err != nil {
			return err
		}
	}
	ww.Flush()
	return nil
}

type Event struct {
	Type            string     `json:"type"`
	Datetime        *time.Time `json:"datetime"`
	HomeTeam        string     `json:"home team"`
	AwayTeam        string     `json:"away team"`
	HomeDivision    string     `json:"home_division"`
	AwayDivision    string     `json:"away_division"`
	Facility        string     `json:"facility"`
	FacilityAddress string     `json:"facility_address"`
	Rink            string     `json:"rink"`
}
