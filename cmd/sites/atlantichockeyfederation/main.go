package main

import (
	"calendar-scrapper/internal/client"
	"calendar-scrapper/pkg/cmdutil"
	"calendar-scrapper/pkg/htmlutil"
	"calendar-scrapper/pkg/parser"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/antchfx/htmlquery"

	jsoniter "github.com/json-iterator/go"
)

const SITE = "atlantichockeyfederation"
const baseUrl = "https://atlantichockeyfederation.com/schedule/?level_id=80"
const seasonId = "131"

var cl = client.GetClient("", 5)

func main() {
	log.Println("started atlantichockeyfederation")
	flags := cmdutil.ParseCommonFlags()

	var mm, yyyy int
	var err error

	today := time.Now()
	mm = int(today.Month())
	yyyy = int(today.Year())

	if *flags.Date != "" {
		mm, yyyy = parser.ParseMonthYear(*flags.Date)
	}

	var games []Game

	divisions, err := fetchDivisions()
	if err != nil {
		log.Fatalf("failed to fetch divisions: %v", err)
	}

	var allGames = [][]string{}
	creds, err := fetchCreds()
	if err != nil {
		log.Fatalf("failed to fetch credentials: %v", err)
	}

	var lock = sync.Mutex{}

	cutoff, err := time.Parse("2006-1-02", strconv.Itoa(yyyy)+"-"+strconv.Itoa(mm)+"-01")
	if err != nil {
		log.Fatalf("failed to parse cutoff date: %v", err)
	}

	var wg = sync.WaitGroup{}
	for _, division := range divisions {
		d := division
		wg.Go(func() {
			games, err = fetchSchedules(d, cutoff, creds)
			log.Printf("division: %s, games: %d", d.name, len(games))
			if err != nil {
				fmt.Printf("error %s\n", err)
				return
			}
			lock.Lock()
			for _, g := range games {
				allGames = append(allGames, []string{
					g.DateTime,
					SITE,
					g.HomeTeam,
					g.AwayTeam,
					g.Location,
					d.name,
					"", // no address
				})
			}
			lock.Unlock()
		})
	}
	wg.Wait()

	log.Println("total games ", len(allGames))
	if *flags.ImportLocations {
		if err := cmdutil.ImportLocations(SITE, allGames); err != nil {
			log.Fatal(err)
		}
	}

	sort.Sort(parser.ByDate(allGames))

	if err := cmdutil.WriteOutput(*flags.Outfile, allGames); err != nil {
		log.Fatal(err)
	}
}

type Creds struct {
	username string
	secret   string
	apiUrl   string
	leagueId string
}

type Game struct {
	Date     string `json:"date"` // 2025-08-29
	Time     string `json:"time"` // 15:00:00
	DateTime string
	HomeTeam string `json:"home_team"`
	AwayTeam string `json:"away_team"`
	Location string `json:"location"`
}

func fetchSchedules(division Division, cutoff time.Time, creds *Creds) ([]Game, error) {
	params := map[string]string{
		"league_id":  creds.leagueId,
		"level_id":   division.id,
		"stat_class": "1",
		"season_id":  seasonId,
	}

	url := signUrl(
		creds.username,
		creds.secret,
		"https://"+creds.apiUrl,
		"get_schedule",
		params,
	)

	resp, err := cl.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get error %s %w", url, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error status code %d for url %s", resp.StatusCode, url)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("http error failed to read body for %s %w", url, err)
	}

	r := map[string][]Game{}

	if err := jsoniter.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	var filtered = []Game{}
	for i := range r["games"] {
		dt, err := time.Parse("2006-01-02 15:04:05", r["games"][i].Date+" "+r["games"][i].Time)
		if err != nil {
			return nil, fmt.Errorf("error parsing date %s %s: %w", r["games"][i].Date, r["games"][i].Time, err)
		}
		if cutoff.After(dt) {
			continue
		}
		r["games"][i].DateTime = dt.Format("2006-01-02 15:04")
		filtered = append(filtered, r["games"][i])
	}
	return filtered, nil
}

func fetchCreds() (*Creds, error) {
	// fmt.Println("fetching creds")
	resp, err := cl.Get(baseUrl)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	body := string(b)

	usernameRe := regexp.MustCompile(`username:\s+"([a-zA-Z0-1.]+)"`)
	secretRe := regexp.MustCompile(`secret:\s+"(\w+)"`)
	apiUrlRe := regexp.MustCompile(`api_url:\s+"([a-zA-Z0-1.]+)"`)
	leagueIdRe := regexp.MustCompile(`league_id:\s+"(\w+)"`)

	username := usernameRe.FindStringSubmatch(body)
	if len(username) == 0 {
		return nil, fmt.Errorf("username not found")
	}
	// fmt.Printf("%+v\n", username)
	secret := secretRe.FindStringSubmatch(body)
	if len(secret) == 0 {
		return nil, fmt.Errorf("secret not found")
	}
	apiUrl := apiUrlRe.FindStringSubmatch(body)
	if len(apiUrl) == 0 {
		return nil, fmt.Errorf("api url not found")
	}
	leagueId := leagueIdRe.FindStringSubmatch(body)
	if len(leagueId) == 0 {
		return nil, fmt.Errorf("league id not found")
	}
	return &Creds{
		username: username[1],
		secret:   secret[1],
		apiUrl:   apiUrl[1],
		leagueId: leagueId[1],
	}, nil
}
func signUrl(username, secret, apiUrl, path string, params map[string]string) string {
	authTimestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// MD5 of empty body (hex)
	bodyMd5 := md5.Sum([]byte(""))
	bodyMd5Hex := hex.EncodeToString(bodyMd5[:])

	// Merge params
	params["auth_key"] = username
	params["auth_timestamp"] = authTimestamp
	params["body_md5"] = bodyMd5Hex

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}

	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, params[k]))
	}
	canonicalQuery := ""
	for i, p := range pairs {
		if i > 0 {
			canonicalQuery += "&"
		}
		canonicalQuery += p
	}

	// String to sign
	stringToSign := fmt.Sprintf("GET\n/%s\n%s", path, canonicalQuery)

	// HMAC-SHA256, hex output
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(stringToSign))
	authSignature := hex.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("%s/%s?%s", apiUrl, path, canonicalQuery+"&auth_signature="+authSignature)
}

type Division struct {
	name string
	id   string
}

func fetchDivisions() ([]Division, error) {
	url := "https://atlantichockeyfederation.com/game-center/"

	resp, err := cl.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http error %w", err)
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("body error %w", err)
	}
	content := string(b)

	doc, err := htmlquery.Parse(strings.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse body")
	}
	if doc == nil {
		return nil, fmt.Errorf("doc nil")
	}

	divisions := []Division{}

	cards := htmlquery.Find(doc, `//div[@class="container"]//div[@class="card"]`)
	if len(cards) == 0 {
		return nil, fmt.Errorf("division cards not found")
	}
	re := regexp.MustCompile("level_id=([0-9]+)")
	for _, card := range cards {
		name := htmlquery.FindOne(card, "h3")
		if name == nil {
			return nil, fmt.Errorf("division node nil")
		}
		divisionName := htmlquery.InnerText(name)
		if divisionName == "" {
			return nil, fmt.Errorf("division name empty")
		}

		// fmt.Println(divisionName)
		for _, link := range htmlquery.Find(card, "//a") {
			url := htmlutil.GetAttr(link, "href")
			if url == "" {
				return nil, fmt.Errorf("division url not found")
			}

			matches := re.FindStringSubmatch(url)
			if len(matches) == 0 {
				return nil, fmt.Errorf("level id not found")
			}
			divisions = append(divisions, Division{
				name: divisionName,
				id:   matches[1],
			})
			break
		}
	}
	return divisions, nil
}
