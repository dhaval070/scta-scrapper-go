package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"calendar-scrapper/internal/webdriver"

	"github.com/antchfx/htmlquery"
	"github.com/tebeka/selenium"
	"golang.org/x/net/html"
)

const SessionID = "8683"

var siteUrl = "https://www.lugsports.com/stats#/1709/schedule?season_id=" + SessionID

var driver selenium.WebDriver
var client = http.DefaultClient

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

		driver = webdriver.GetWebDriver()

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

	if *infile == "" {
		var events []Event
		lastEvent := result[len(result)-1]
		val, err := driver.ExecuteScript("return window.localStorage.getItem('website_api_ticket')", nil)
		if err != nil {
			log.Println("failed to get api token")
		}
		ticket := val.(string)

		for i := 0; i < 50; i += 1 {
			events, err = getMore(lastEvent.ID, ticket)
			if err != nil {
				panic(err)
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
		fmt.Println("total ", len(result))
	}

	if *flags.ImportLocations {
		if err := cmdutil.ImportLocations(SITE, result); err != nil {
			log.Fatal(err)
		}
	}

	if *flags.Outfile != "" {
		fh, err := os.Create(*flags.Outfile)
		if err != nil {
			log.Fatal(err)
		}
		WriteEvents(fh, result)
		fh.Close()
	}
}

func getMore(lastID string, ticket string) ([]Event, error) {
	s := "https://web.api.digitalshift.ca/partials/stats/schedule/table?order=datetime&season_id=" + SessionID + "&start_id=" + lastID + "&offset=1&limit=200&all=true"

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
