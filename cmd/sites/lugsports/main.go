package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
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
)

var sites = [][]string{
	{"https://www.lugsports.com/stats#/1709/schedule?season_id=", "8683"},
	{"https://www.lugsports.com/stats#/1869/schedule?season_id=", "9203"},
	{"https://www.lugsports.com/stats#/162/schedule?season_id=", "9350"},
}

var driver selenium.WebDriver
var client = http.DefaultClient

const SITE = "lugsports"

func main() {
	importLocations := flag.Bool("import-locations", false, "import site locations")
	outfile := flag.String("outfile", "", "output filename")
	// date flag is not used but have to add here to make this command compatible with other sites and use with run.sh and run-all.sh
	_ = flag.String("date", "", "calendar month and year in format: mmyyyy")

	flag.Parse()

	var err error

	service, err := selenium.NewChromeDriverService("./chromedriver", 4444)
	if err != nil {
		log.Fatal("Error:", err)
	}
	defer service.Stop()

	driver = webdriver.GetWebDriver()

	var result []Event
	for _, site := range sites {
		r, err := scrapeSeason(driver, site[0], site[1])
		if err != nil {
			log.Printf("lugsports error in %s, %s: %s\n", site[0], site[1], err.Error())
			continue
		}
		result = append(result, r...)
	}

	fmt.Println("total ", len(result))

	if *importLocations {
		config.Init("config", ".")
		cfg := config.MustReadConfig()

		var locations = make([]model.SitesLocation, 0, len(result))
		for _, r := range result {
			l := model.SitesLocation{
				Location: r.Facility + "(" + r.Rink + ")",
				Address:  r.FacilityAddress,
				// surface will be populated later by extracting from Location field
				// Surface:  r.Rink,
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

func scrapeSeason(driver selenium.WebDriver, siteUrl, seasonId string) (events []Event, err error) {
	err = driver.Get(siteUrl + seasonId)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s: %v\n", siteUrl, err)
	}

	time.Sleep(10 * time.Second)
	source, err := driver.PageSource()

	if err != nil {
		return nil, fmt.Errorf("pageSource error %s: %v\n", siteUrl, err)
	}

	doc, err := htmlquery.Parse(strings.NewReader(source))
	if err != nil {
		return nil, errors.New("failed to parse html")
	}

	node := htmlquery.FindOne(doc, `//partial[@slug="stats/schedule/table"]/div//table[@class="schedule"]/tbody`)
	if node == nil {
		return events, errors.New("schedule not found")
	}

	var content string
	for _, attr := range node.Attr {
		if attr.Key == "ng-init" {
			content = attr.Val
			break
		}
	}
	if content == "" {
		return nil, errors.New("ng-init not found")
	}
	content = strings.Replace(content, "ctrl.schedule=", "", 1)

	var result []Event

	if err = json.Unmarshal([]byte(content), &result); err != nil {
		return nil, errors.New("failed to unmarshal schedule json ")
	}

	for i, v := range result {
		address, err := url.QueryUnescape(v.FacilityAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to unescape address %s ", v.FacilityAddress)
		}
		result[i].FacilityAddress = address
	}

	if len(result) == 0 {
		return nil, errors.New("no results found")
	}

	lastEvent := result[len(result)-1]
	val, err := driver.ExecuteScript("return window.localStorage.getItem('website_api_ticket')", nil)
	if err != nil {
		return nil, errors.New("failed to get api token")
	}
	ticket := val.(string)

	for i := 0; i < 50; i += 1 {
		events, err = getMore(seasonId, lastEvent.ID, ticket)
		if err != nil {
			return nil, err
		}
		if len(events) == 0 {
			break
		}
		// b, err := json.Marshal(events)
		// if err != nil {
		// 	panic(err)
		// }
		// fmt.Println(string(b))
		result = append(result, events...)
		lastEvent = events[len(events)-1]
	}
	return result, nil
}

func getMore(seasonId, lastID, ticket string) ([]Event, error) {
	s := "https://web.api.digitalshift.ca/partials/stats/schedule/table?order=datetime&season_id=" + seasonId + "&start_id=" + lastID + "&offset=1&limit=200&all=true"

	req, err := http.NewRequest("GET", s, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", `ticket="`+ticket+`"`)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Referer", "https://www.lugsports.com/")
	req.Header.Add("Origin", "https://www.lugsports.com/")
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:139.0) Gecko/20100101 Firefox/139.0")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Site", "cross-site")
	req.Header.Add("Sec-GPC", "1")
	req.Header.Add("DNT", "1")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// fmt.Println(string(b))
	var data = map[string][]Event{}
	err = json.Unmarshal(b, &data)
	if err != nil {
		log.Println(string(b))
		return nil, err
	}
	for i, v := range data["schedule"] {
		address, err := url.QueryUnescape(v.FacilityAddress)
		if err != nil {
			panic(fmt.Errorf("failed to unescape address %s ", v.FacilityAddress))
		}
		data["schedule"][i].FacilityAddress = address
	}
	return data["schedule"], nil
}

func WriteEvents(w io.Writer, data []Event) error {
	var err error
	ww := csv.NewWriter(w)
	for _, rec := range data {
		row := []string{
			rec.Datetime.Format("2006-1-02 15:04"),
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
	ID              string     `json:"id"`
	Type            string     `json:"type"`
	Datetime        *time.Time `json:"datetime"`
	HomeTeam        string     `json:"home_team"`
	AwayTeam        string     `json:"away_team"`
	HomeDivision    string     `json:"home_division"`
	AwayDivision    string     `json:"away_division"`
	Facility        string     `json:"facility"`
	FacilityAddress string     `json:"facility_address"`
	Rink            string     `json:"rink"`
}
