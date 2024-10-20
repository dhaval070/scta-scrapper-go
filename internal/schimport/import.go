package schimport

import (
	"calendar-scrapper/dao/model"
	"calendar-scrapper/pkg/repository"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

/*
	 {
		  "status": true,
		  "games": [
			{
			  "game_id": "542257",
			  "rink": "York 3",
			  "start_time": "09:10",
			  "start_date": "2024-10-27",
			  "division": "U10",
			  "category": "A",
			  "visitor": "The Attack",
			  "visitor_team_id": "HC 20241049800001451",
			  "home": "MHL-Applewood Coyotes",
			  "home_team_id": "HC 2024141800001442",
			  "game_type": "LG"
			},
			]
		}
*/
var client = http.Client{}

type Importer struct {
	repo   *repository.Repository
	apiKey string
	url    string
}

func NewImporter(repo *repository.Repository, apiKey, url string) *Importer {
	return &Importer{
		repo:   repo,
		apiKey: apiKey,
		url:    url,
	}
}

type Data struct {
	Games []Game
}

type Game struct {
	GameID        string `json:"game_id"`
	Rink          string `json:"rink"`
	StartTime     string `json:"start_time"`
	StartDate     string `json:"start_date"`
	Division      string `json:"division"`
	Category      string `json:"category"`
	Visitor       string `json:"visitor"`
	VisitorTeamID string `json:"visitor_team_id"`
	Home          string `json:"home"`
	HomeTeamID    string `json:"home_team_id"`
	GameType      string `json:"game_type"`
}

func (i *Importer) FetchAndImport(site string, m map[string]int, cdate time.Time) error {
	var err error

	b, err := i.FetchJson(site, cdate)
	if err != nil {
		return err
	}

	var data Data
	if err = json.Unmarshal(b, &data); err != nil {
		return err
	}

	if len(data.Games) == 0 {
		log.Println("no games to import")
		return nil
	}
	return i.ImportJson(site, data, cdate, m)
}

func (i *Importer) ImportJson(site string, data Data, cutOffDate time.Time, mapping map[string]int) error {
	log.Println("importing json")

	var err error
	var SourceType = "file"

	m := []*model.Event{}

	ww := csv.NewWriter(os.Stdout)
	var r = make([]string, 11)

	mappingUpdates := map[string]int32{}

	for _, g := range data.Games {
		dt, err := parseDate("2006-01-02", g.StartDate, g.StartTime)
		if err != nil {
			log.Printf("error: failed parsing: date: %s , time: %s", g.StartDate, g.StartTime)
			continue
		}

		if dt.Before(cutOffDate) {
			continue
		}

		sid, ok := mapping[g.Rink]
		if !ok || sid == 0 {
			mappingUpdates[g.Rink] = 0
			log.Printf("failed to map surfaceid %s\n", g.Rink)
			continue
		}

		m = append(m, &model.Event{
			Site:        site,
			SourceType:  SourceType,
			Datetime:    dt,
			HomeTeam:    g.Home,
			OidHome:     g.HomeTeamID,
			GuestTeam:   g.Visitor,
			OidGuest:    g.VisitorTeamID,
			Location:    g.Rink,
			Division:    g.Division + " " + g.Category,
			SurfaceID:   int32(sid),
			DateCreated: time.Now(),
		})

		r[0] = g.GameID
		r[1] = g.Rink
		r[2] = g.StartTime
		r[3] = g.StartDate
		r[4] = g.Division
		r[5] = g.Category
		r[6] = g.Visitor
		r[7] = g.VisitorTeamID
		r[8] = g.Home
		r[9] = g.HomeTeamID

		r[7] = strings.Replace(r[7], "HC ", "", -1)
		r[9] = strings.Replace(r[9], "HC ", "", -1)
		r[3] = dt.Format("2006-01-02")
		r[10] = fmt.Sprint(sid)
		ww.Write(r)
	}
	ww.Flush()

	log.Println("total events ", len(m))

	err = i.repo.ImportEvents(site, m, cutOffDate)
	if err != nil {
		return err
	}
	return i.repo.ImportMappings(site, mappingUpdates)
}

func parseDate(dateFormat, date, t string) (tt time.Time, err error) {
	parts := strings.Split(t, ":")
	if len(parts[0]) < 2 {
		parts[0] = "0" + parts[0]
	}

	if strings.Index(t, "PM") > -1 {
		h, err := strconv.ParseInt(parts[0], 10, 32)
		if err != nil {
			return tt, fmt.Errorf("failed to convert hours to int %w", err)
		}
		h += 12
		parts[0] = fmt.Sprintf("%02d", h)
		re := regexp.MustCompile(`\s*PM`)
		parts[1] = re.ReplaceAllString(parts[1], "")
	}

	tj := strings.Join(parts, ":")

	t1, err := time.Parse("15:04", tj)
	if err != nil {
		return tt, fmt.Errorf("failed to parse time %s", t)
	}

	// 01-Oct-2023
	dt, err := time.Parse(dateFormat+" 15:04:05", date+" "+t1.Format("15:04:05"))

	if err != nil {
		return tt, fmt.Errorf("failed to parse date:%s time:%s :%w", date, t, err)
	}
	return dt, nil
}

// Game_id	Rink	StartTime	StartDate	Division	Category	Visitor	VisitorTeamID	Home	HomeTeamID	GameType
func (i *Importer) Importxls(site string, root *html.Node, cutOffDate time.Time, mapping map[string]int) error {
	var err error
	var SourceType = "file"

	m := []*model.Event{}

	rows, err := htmlquery.QueryAll(root, "//table/tbody/tr")
	if err != nil {
		return err
	}

	ww := csv.NewWriter(os.Stdout)

	var r = make([]string, 11)

	for _, row := range rows[1:] {
		cols, err := htmlquery.QueryAll(row, "//td")
		if err != nil {
			return err
		}

		dt, err := parseDate("02-Jan-2006", htmlquery.InnerText(cols[3]), htmlquery.InnerText(cols[2]))
		if err != nil {
			return err
		}

		if dt.Before(cutOffDate) {
			continue
		}

		sid, ok := mapping[htmlquery.InnerText(cols[1])]
		if !ok {
			log.Printf("failed to map surfaceid %s\n", htmlquery.InnerText(cols[1]))
			continue
		}

		m = append(m, &model.Event{
			Site:        site,
			SourceType:  SourceType,
			Datetime:    dt,
			HomeTeam:    htmlquery.InnerText(cols[8]),
			OidHome:     htmlquery.InnerText(cols[9]),
			GuestTeam:   htmlquery.InnerText(cols[6]),
			OidGuest:    htmlquery.InnerText(cols[7]),
			Location:    htmlquery.InnerText(cols[1]),
			Division:    htmlquery.InnerText(cols[4]) + " " + htmlquery.InnerText(cols[5]),
			SurfaceID:   int32(sid),
			DateCreated: time.Now(),
		})

		for i := 0; i < 10; i += 1 {
			r[i] = htmlquery.InnerText(cols[i])
		}
		r[7] = strings.Replace(r[7], "HC ", "", -1)
		r[9] = strings.Replace(r[9], "HC ", "", -1)
		r[3] = dt.Format("2006-01-02")
		r[10] = fmt.Sprint(sid)

		ww.Write(r)
	}
	ww.Flush()

	log.Println("total events ", len(m))
	err = i.repo.ImportEvents(site, m, cutOffDate)
	return err
}

func (i *Importer) FetchJson(site string, cdate time.Time) ([]byte, error) {
	log.Println("fetching json")

	url := fmt.Sprintf(i.url, cdate.Format("02-Jan-2006"), site)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("x-api-key", i.apiKey)

	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	return b, err
}
