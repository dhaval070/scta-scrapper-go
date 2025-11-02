package parser

import (
	"calendar-scrapper/internal/client"
	"calendar-scrapper/pkg/month"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

var Client = client.GetClient(os.Getenv("HTTP_PROXY"))

type ByDate [][]string

func (a ByDate) Len() int {
	return len(a)
}

func (a ByDate) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByDate) Less(i, j int) bool {
	return a[i][0] < a[j][0]
}

func GetAttr(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func ParseTime(content string) (string, error) {
	reg := regexp.MustCompile("([0-9]{1,2}):([0-9]{1,2}) (AM|PM)")

	parts := reg.FindStringSubmatch(content)

	if parts == nil {
		return "", fmt.Errorf("failed to parse time %s", content)
	}

	h, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", errors.New("failed to convert hour")
	}

	m, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", errors.New("failed to convert minutes")
	}

	if parts[3] == "PM" && h < 12 {
		h += 12
	}

	return fmt.Sprintf("%02d:%02d", h, m), nil
}

func ParseSchedules(site string, doc *html.Node) [][]string {
	nodes := htmlquery.Find(doc, `//div[contains(@class, "day-details")]`)

	var result = [][]string{}

	var lock = &sync.Mutex{}
	var wg = &sync.WaitGroup{}

	for _, node := range nodes {
		id := GetAttr(node, "id")
		ymd := ParseId(id)

		listItems := htmlquery.Find(node, `//div[contains(@class, "event-list-item")]/div`) // `div[2]`)

		for _, parent := range listItems {
			item := htmlquery.FindOne(parent, `div[2]`)
			content := htmlquery.OutputHTML(item, true)

			if strings.Contains(content, "All Day") || strings.Contains(content, "time-secondary") || strings.Contains(content, "Cancelled") {
				continue
			}
			timeval, err := ParseTime(content)
			if err != nil {
				log.Fatal(err, content)
				continue
			}

			division, err := QueryInnerText(item, `//span[@class="game_no"]`)
			if err != nil {
				log.Println(err)
				continue
			}
			guestTeam, err := QueryInnerText(item, `//div[contains(@class, "subject-owner")]`)
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
			location, err := QueryInnerText(item, `//div[@class="location remote"]`)
			if err != nil {
				log.Println("failed to parse location")
				continue
			}

			item = htmlquery.Find(parent, `div[1]//a[@class="remote" or @class="local"]`)[0]
			var url string
			var address string
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
				go func(url string, wg *sync.WaitGroup, lock *sync.Mutex) {
					defer wg.Done()
					address = GetVenueAddress(url, class)
					lock.Lock()
					result = append(result, []string{ymd + " " + timeval, site, homeTeam, guestTeam, location, division, address})
					lock.Unlock()
				}(url, wg, lock)
			} else {
				result = append(result, []string{ymd + " " + timeval, site, homeTeam, guestTeam, location, division, address})
			}
		}
	}
	wg.Wait()
	return result
}

func QueryInnerText(doc *html.Node, expr string) (string, error) {
	node, err := htmlquery.Query(doc, expr)
	if err != nil {
		return "", err
	}
	if node != nil {
		return htmlquery.InnerText(node), nil
	}
	return "", fmt.Errorf("node not found %v", expr)
}

func ParseId(id string) string {
	parts := strings.Split(id, "-")

	if len(parts) != 4 {
		log.Fatal("len not 4")
	}
	mm, err := month.MonthString(parts[1])

	if err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%s-%02d-%s", parts[3], mm, parts[2])
}

func GetVenueAddress(url string, class string) string {
	req, err := http.NewRequest("GET", url, nil)
	resp, err := Client.Do(req)

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

	var address string

	// theonedb.com
	if class == "remote" {
		item := htmlquery.FindOne(doc, `//div[@class="container"]/div/div/h2/small[2]`)
		if item == nil {
			log.Println("address node not found, url:", url)
			return ""
		}
		address = htmlquery.InnerText(item)
	} else if class == "local" {
		// local url
		node := htmlquery.FindOne(doc, `//div[@class="month"]/following-sibling::div/div/div`)
		if node == nil {
			log.Println("address node not found, url:", url)
			return ""
		}
		address = htmlquery.InnerText(node)
	}
	log.Println(url + ":" + address)
	return address
}

func ParseMonthYear(dt string) (int, int) {
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

func FetchSchedules(site, url string, groups map[string]string, mm, yyyy int) [][]string {

	var schedules = make([][]string, 0)

	for division, id := range groups {
		url := fmt.Sprintf(url, id, mm, yyyy)
		doc, err := htmlquery.LoadURL(url)
		if err != nil {
			log.Fatal("load calendar url", err)
		}

		result := ParseSchedules(site, doc)

		for _, row := range result {
			row[5] = division
			schedules = append(schedules, row)
		}
	}

	sort.Sort(ByDate(schedules))
	return schedules
}

// DayDetailsConfig configures the ParseDayDetailsSchedule function
type DayDetailsConfig struct {
	TournamentCheckExact bool                // true for exact match, false for contains check
	LogErrors            bool                // enable verbose error logging
	GameDetailsFunc      func(string) string // function to fetch venue address from game URL
}

// ParseDayDetailsSchedule parses schedules from day-details divs with home/away logic
// This is used by sites like ckgha, bghc, cygha, georginagirlshockey, lakeshorelightning,
// londondevilettes, pgha, sarniagirlshockey, scarboroughsharks, smgha, wgha
func ParseDayDetailsSchedule(doc *html.Node, site, baseURL, homeTeam string, cfg DayDetailsConfig) [][]string {
	nodes := htmlquery.Find(doc, `//div[contains(@class, "day-details")]`)

	var result = [][]string{}
	var lock = &sync.Mutex{}
	var wg = &sync.WaitGroup{}

	for _, node := range nodes {
		id := GetAttr(node, "id")

		if id == "" {
			log.Fatal("id not found")
		}

		ymd := ParseId(id)
		listItems := htmlquery.Find(node, `//div[contains(@class, "event-list-item")]/div`)

		for _, parent := range listItems {
			item := htmlquery.FindOne(parent, `div[2]`)
			content := htmlquery.OutputHTML(item, true)

			var homeGame = true
			var gameType = htmlquery.InnerText(htmlquery.FindOne(item, `div[2]`))

			// Skip tournaments
			if cfg.TournamentCheckExact {
				if gameType == "Tournament" || gameType == "Hosted Tournament" {
					continue
				}
			} else {
				if strings.Contains(gameType, "Tournament") {
					continue
				}
			}

			// Skip practices
			if strings.Contains(gameType, "Practice") {
				continue
			}

			// Determine if away game
			if strings.Contains(gameType, "Away") {
				homeGame = false
			}

			timeval, err := ParseTime(content)
			if err != nil {
				if cfg.LogErrors {
					log.Println(err)
				}
				continue
			}

			division, err := QueryInnerText(item, `div[3]/div[1]`)
			if err != nil {
				if cfg.LogErrors {
					log.Println(err)
				}
				continue
			}

			subjectText, err := htmlquery.Query(item, `//div[contains(@class, "subject-text")]`)
			if err != nil {
				if cfg.LogErrors {
					log.Println(err)
				}
				continue
			}

			ch := subjectText.FirstChild
			guestTeam := strings.Replace(htmlquery.InnerText(ch), "@ ", "", 1)
			guestTeam = strings.Replace(guestTeam, "vs ", "", 1)

			hm := homeTeam
			if !homeGame {
				hm, guestTeam = guestTeam, homeTeam
			}

			location, err := QueryInnerText(item, `//div[contains(@class,"location")]`)

			item = htmlquery.FindOne(parent, `//div[1]/div[2]/a`)
			url := GetAttr(item, "href")

			if url != "" {
				wg.Add(1)
				if url[:4] != "http" {
					url = baseURL + url
				}

				go func(url string, location string, wg *sync.WaitGroup, lock *sync.Mutex) {
					defer wg.Done()
					address := cfg.GameDetailsFunc(url)
					if address == "" {
						log.Println("addr not found ", url)
					} else {
						log.Println(url, address)
					}
					lock.Lock()
					result = append(result, []string{ymd + " " + timeval, site, hm, guestTeam, location, division, address})
					lock.Unlock()
				}(url, location, wg, lock)
			} else {
				log.Println("url not found")
				result = append(result, []string{ymd + " " + timeval, site, hm, guestTeam, location, division, ""})
			}
		}
	}
	wg.Wait()
	return result
}

// MonthScheduleConfig configures the ParseMonthBasedSchedule function
type MonthScheduleConfig struct {
	TeamParseStrategy string                      // "subject-owner-first" or "first-char-detect"
	URLPrefix         string                      // optional base URL prefix
	VenueAddressFunc  func(string, string) string // function to fetch venue address
}

// ParseMonthBasedSchedule parses schedules using month/year parameters
// NOTE: Currently unused - created as a template for future month-based sites
// Existing sites (heoaaaleague, wmha, windsoraaazone, spfhahockey) have different
// implementations and would require manual adaptation and testing to use this function
func ParseMonthBasedSchedule(doc *html.Node, mm, yyyy int, site string, cfg MonthScheduleConfig) [][]string {
	nodes := htmlquery.Find(doc, `//div[contains(@class, "day-details")]`)

	var result = [][]string{}
	var lock = &sync.Mutex{}
	var wg = &sync.WaitGroup{}

	for _, node := range nodes {
		listItems := htmlquery.Find(node, `//div[contains(@class, "event-list-item")]/div`)
		for _, parent := range listItems {
			items := htmlquery.Find(parent, `div[2]`)
			if len(items) == 0 {
				continue
			}
			item := items[0]
			content := htmlquery.OutputHTML(item, true)

			if strings.Contains(content, "All Day") || strings.Contains(content, "time-secondary") || strings.Contains(content, "Cancelled") {
				continue
			}

			timeval, err := ParseTime(content)
			if err != nil {
				log.Println(err)
				continue
			}

			txt, err := QueryInnerText(item, `//div[@class="day_of_month"]`)
			if err != nil {
				log.Println(err)
				continue
			}
			dom := txt[4:]
			ymd := fmt.Sprintf("%d-%d-%s", yyyy, mm, dom)

			var division, homeTeam, guestTeam string

			if cfg.TeamParseStrategy == "first-char-detect" {
				// Used by wmha
				division, err = QueryInnerText(item, `//div[contains(@class,"subject-owner")]`)
				if err != nil {
					log.Println("subject owner error ", err, content)
					continue
				}

				subjectText, err := htmlquery.Query(item, `//div[contains(@class, "subject-text")]`)
				if err != nil {
					log.Println(err)
					continue
				}

				ch := htmlquery.InnerText(subjectText.FirstChild)
				if len(ch) > 0 && ch[0] == '@' {
					guestTeam = site // Assume site has HOME_TEAM value
					homeTeam = ch[2:]
				} else if len(ch) > 2 && ch[0:3] == "vs " {
					homeTeam = site
					guestTeam = ch[3:]
				} else {
					log.Println("failed to parse teams from:", ch)
					continue
				}
			} else {
				// Default: "subject-owner-first" - used by heoaaaleague, etc
				division, err = QueryInnerText(item, `//span[@class="game_no"]`)
				if err != nil {
					log.Println(err)
					continue
				}
				guestTeam, err = QueryInnerText(item, `//div[contains(@class, "subject-owner")]`)
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
				homeTeam = strings.Replace(htmlquery.InnerText(ch), "@ ", "", -1)
			}

			location, err := QueryInnerText(item, `//div[@class="location remote"]`)
			if err != nil {
				// Try alternative location class
				location, _ = QueryInnerText(item, `//div[contains(@class,"location")]`)
			}

			items = htmlquery.Find(parent, `div[1]//a[@class="remote" or @class="local"]`)
			if len(items) == 0 {
				result = append(result, []string{ymd + " " + timeval, site, homeTeam, guestTeam, location, division, ""})
				continue
			}
			item = items[0]

			var url string
			var address string
			var class string

			for _, attr := range item.Attr {
				if attr.Key == "href" {
					url = attr.Val
				} else if attr.Key == "class" {
					class = attr.Val
				}
			}

			if url != "" {
				if cfg.URLPrefix != "" && url[:4] != "http" {
					url = cfg.URLPrefix + url
				}
				wg.Add(1)
				go func(url string, class string, wg *sync.WaitGroup, lock *sync.Mutex) {
					defer wg.Done()
					address = cfg.VenueAddressFunc(url, class)
					lock.Lock()
					result = append(result, []string{ymd + " " + timeval, site, homeTeam, guestTeam, location, division, address})
					lock.Unlock()
				}(url, class, wg, lock)
			} else {
				result = append(result, []string{ymd + " " + timeval, site, homeTeam, guestTeam, location, division, address})
			}
		}
	}
	wg.Wait()
	return result
}

// ParseSiteListGroups extracts division groups from site-list divs
// Used by 17 sites: beechey, eomhl, essexll, fourcountieshockey, gbmhl, gbtll,
// grandriverll, haldimandll, intertownll, leohockey, lmll, ndll,
// omha-aaa, srll, threecountyhockey, victoriadurham, woaa.on
func ParseSiteListGroups(doc *html.Node, xpath string) map[string]string {
	links := htmlquery.Find(doc, xpath)
	var groups = make(map[string]string)

	for _, n := range links {
		href := GetAttr(n, "href")
		division := htmlquery.InnerText(n)

		re := regexp.MustCompile(`Groups/(.+)/`)
		parts := re.FindAllStringSubmatch(href, -1)
		if parts == nil {
			log.Fatal("failed to parse group link", href)
		}
		groups[division] = parts[0][1]
	}
	return groups
}

// GetVenueDetailsFromRelativeURL fetches venue address from a relative URL
// Used by 16 sites as venueDetails() function
func GetVenueDetailsFromRelativeURL(baseURL, relativeURL string) string {
	resp, err := Client.Get(baseURL + relativeURL)
	if err != nil {
		log.Println("error get "+relativeURL, err)
		return ""
	}
	defer resp.Body.Close()

	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		log.Println("error parsing "+relativeURL, err)
		return ""
	}

	node := htmlquery.FindOne(doc, `//div[@class="month"]/following-sibling::div/div/div`)
	if node == nil {
		log.Println("error venue detail not found ", baseURL+relativeURL)
		return ""
	}
	address := htmlquery.InnerText(node)
	return address
}

// GetGameDetailsAddress finds and fetches venue address from game details page
// Used by 16 sites as gameDetails() function
// Looks for "More Venue Details" link and fetches the venue address
func GetGameDetailsAddress(gameURL, baseURL string) string {
	resp, err := Client.Get(gameURL)
	if err != nil {
		log.Println(err)
		return ""
	}
	defer resp.Body.Close()

	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		log.Println("error getting "+gameURL, err)
		return ""
	}

	var venueURL string
	nodes := htmlquery.Find(doc, `//div[contains(@class,"game-details-tabs")]//a`)

	for _, n := range nodes {
		venueURL = GetAttr(n, "href")

		if htmlquery.InnerText(n) == "More Venue Details" {
			return GetVenueDetailsFromRelativeURL(baseURL, venueURL)
		}
	}
	return ""
}
