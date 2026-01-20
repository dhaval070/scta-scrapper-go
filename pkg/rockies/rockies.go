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
	"strconv"
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

	nodes := htmlquery.Find(doc, `//h4[@class="panel-title"]/a`)
	// log.Println("nodes ", len(nodes))
	if len(nodes) == 0 {
		return nil, errors.New("cannot find divisions links")
	}

	re := regexp.MustCompile("[0-9]+")
	allGames := []Game{}
	lock := sync.Mutex{}
	var wg sync.WaitGroup
	for _, node := range nodes {
		dataParent := htmlutil.GetAttr(node, "data-parent")
		href := htmlutil.GetAttr(node, "href")

		if dataParent == "" {
			return nil, errors.New("data parent not found")
		}
		if href == "" {
			return nil, errors.New("href not found")
		}
		parent := re.FindString(dataParent)
		if parent == "" {
			return nil, errors.New("parent id not found")
		}
		child := re.FindString(href)
		if child == "" {
			return nil, errors.New("child path id not found")
		}

		wg.Add(1)
		go func(parent, child string) {
			defer wg.Done()
			games, err := r.fetchRockiesGames(parent, child, yyyy)
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
		dt, err := time.Parse("2006-01-02T15:04:05", g.EDate)
		if err != nil {
			log.Println("invalid date format: ", g.EDate)
			continue
		}
		if dt.Year() > yyyy || int(dt.Month()) > mm {
			continue
		}
		result = append(result, []string{
			dt.Format("2006-01-02 15:04"),
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

func (rs *RockiesScraper) fetchRockiesGames(parent, child string, yyyy int) (result []Game, err error) {
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
		result = append(result, games...)
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
		matches := reSeason.FindStringSubmatch(val)
		if len(matches) == 0 {
			return "", fmt.Errorf("failed to parse season %s", val)
		}
		log.Printf("%+v\n", matches)
		fromInt, _ := strconv.Atoi(matches[1])
		toInt, _ := strconv.Atoi(matches[2])

		if yyyy >= fromInt && yyyy <= toInt {
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
