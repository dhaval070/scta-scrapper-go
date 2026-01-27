// this is scraper package for sites rockieshockeyleague.com
package rockies

import (
	"bytes"
	"calendar-scrapper/pkg/htmlutil"
	"calendar-scrapper/pkg/parser"
	"calendar-scrapper/pkg/siteconfig"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/antchfx/htmlquery"
	jsoniter "github.com/json-iterator/go"
)

// https://rockieshockeyleague.com/division/1724/15639/games
// GET https://rockieshockeyleague.com/api/leaguegame/get/2155/ 12605/ 1724/15639/  4553/     0/
var client = parser.Client
var reSeason = regexp.MustCompile(`([0-9]+)\s*-\s*([0-9]+)`)

type RockiesScraper struct {
	Sc        *siteconfig.SiteConfig
	ParserCfg *siteconfig.ParserConfigJSON
}

func (r *RockiesScraper) ScrapeRockies(mm, yyyy int) (result [][]string, err error) {
	resp, err := client.Get(r.Sc.BaseURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	doc, err := htmlquery.Parse(bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("failed to parse doc %w", err)
	}

	reDiv := regexp.MustCompile("/division/([0-9]+)/([0-9]+)/games")

	allGames := []Game{}
	lock := sync.Mutex{}
	var wg sync.WaitGroup

	for _, node := range htmlquery.Find(doc, `//div[contains(@class, "panel")]//a]`) {
		href := htmlutil.GetAttr(node, "href")

		matches := reDiv.FindStringSubmatch(href)
		if len(matches) != 3 {
			continue
		}
		parent := matches[1]
		child := matches[2]

		log.Println(parent, child)
		wg.Add(1)
		go func(parent, child string) {
			defer wg.Done()
			games, err := r.fetchRockiesGames(parent, child, mm, yyyy)
			if err != nil {
				log.Printf("scrape error: %v", err)
				return
			}

			ch := make(chan Game, len(games))
			for _, g := range games {
				go getAddress(g, ch)
			}

			var collected = make([]Game, 0, len(games))

			for i := 0; i < len(games); i += 1 {
				collected = append(collected, <-ch)
			}

			lock.Lock()
			allGames = append(allGames, collected...)
			lock.Unlock()
		}(parent, child)
	}
	// wait for all fetches to complete
	wg.Wait()
	log.Println("total ", len(allGames))
	for _, g := range allGames {
		log.Println(g.HomeDivision)
	}

	for _, g := range allGames {
		result = append(result, []string{
			g.Date.Format("2006-01-02 15:04"),
			r.Sc.SiteName,
			g.HomeTeamName,
			g.AwayTeamName,
			g.ArenaName,
			g.HomeDivision,
			g.Address,
		})
	}
	return result, err
}

func (rs *RockiesScraper) fetchRockiesGames(parent, child string, mm, yyyy int) (result []Game, err error) {
	url := rs.Sc.BaseURL + "/division/" + parent + "/" + child + "/games"
	r, err := parser.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	body := string(b)
	re := regexp.MustCompile("getMonthYears/([0-9]+)")

	matches := re.FindStringSubmatch(body)
	if len(matches) == 0 {
		return nil, errors.New("regxp match failed")
	}

	season, err := scrapeSeason(body, yyyy)
	if err != nil {
		return result, err
	}
	if season == "" {
		return result, err
	}

	fromDt := fmt.Sprintf("%d%0d", yyyy, mm)

	for _, gt := range rs.ParserCfg.GameType {
		url = fmt.Sprintf(rs.Sc.BaseURL+"/api/leaguegame/get/%s/%s/%s/%s/%s/0", matches[1], season, parent, child, gt)

		log.Println("url: ", url)
		resp, err := client.Get(url)
		if err != nil {
			return result, err
		}
		defer resp.Body.Close()
		b, err = io.ReadAll(resp.Body)
		if err != nil {
			return result, err
		}
		var games []Game
		err = jsoniter.Unmarshal(b, &games)
		if err != nil {
			log.Println("unmarshal error ", url, err)
			continue
		}

		for i := 0; i < len(games); i += 1 {
			dt, err := time.Parse("2006-01-02T15:04:05", games[i].EDate)
			if err != nil {
				log.Println("invalid date format: ", games[i].EDate)
				continue
			}
			if fromDt > dt.Format("200601") {
				continue
			}
			games[i].Date = dt
			result = append(result, games[i])
		}
	}
	return result, err
}

func scrapeSeason(body string, yyyy int) (string, error) {
	doc, err := htmlquery.Parse(strings.NewReader(body))
	if err != nil {
		return "", err
	}

	for _, node := range htmlquery.Find(doc, `//select[@id="ddlSeason"]/option`) {
		val := htmlquery.InnerText(node)
		if strings.Contains(val, fmt.Sprint(yyyy)) {
			return htmlutil.GetAttr(node, "value"), nil
		}
	}

	log.Println("required season not found")
	return "", nil
}

func getAddress(g Game, ch chan<- Game) {
	defer func() { ch <- g }()

	url := fmt.Sprintf("http://rinkdb.com/v2/view/%s/%s/%s", g.Country, g.Prov, g.RARIDString)
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("http error for %s %s", url, err.Error())
		return
	}
	defer resp.Body.Close()
	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		log.Printf("error parsing doc %s %s", url, err.Error())
		return
	}
	nodes := htmlquery.Find(doc, `//div[@class="container-fluid"]//h3`)
	if len(nodes) == 0 {
		log.Printf("address nodes not found %s", url)
		return
	}
	address := ""
	for _, node := range nodes {
		address = address + htmlquery.InnerText(node) + " "
	}
	log.Println("address: ", address)
	g.Address = strings.Trim(address, " \n")
}
