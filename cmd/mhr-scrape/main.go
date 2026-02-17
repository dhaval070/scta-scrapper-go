package main

import (
	"bytes"
	"calendar-scrapper/config"
	"calendar-scrapper/dao/model"
	"calendar-scrapper/internal/client"
	"calendar-scrapper/pkg/htmlutil"
	"calendar-scrapper/pkg/repository"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/antchfx/htmlquery"
	"github.com/spf13/cobra"
	"golang.org/x/net/html"
	"gorm.io/gorm"
)

var rootCmd = &cobra.Command{
	Use:   "mhr-scrape",
	Short: "Myhockeyrankings scraper and location matcher with livebarn",
	Long:  `A CLI tool to scrape myhockeyrankings venues data and store it in a database. Match venues with livebarn locations`,
	Run:   runScraper,
}

var matchCmd = &cobra.Command{
	Use:   "match",
	Short: "Match mhr venues with livebarn locations",
	Run: func(cmd *cobra.Command, args []string) {
		runMatcher()
	},
}

var cfg config.Config
var repo *repository.Repository
var rePostalCode = regexp.MustCompile(` ([0-9]{5,})\s*| ([a-zA-Z0-9]{3}\s?[a-zA-Z0-9]{3})$`)

const BASE_URL = "https://myhockeyrankings.com"

var cl = client.GetClient("", 5)

func init() {
	rootCmd.AddCommand(matchCmd)
	config.Init("config", ".")

	cfg = config.MustReadConfig()
	repo = repository.NewRepository(cfg)
}

type MhrLocation struct {
	MhrID              int                 `gorm:"column:mhr_id;primaryKey" json:"mhr_id"`
	RinkName           string              `gorm:"column:rink_name;not null" json:"rink_name"`
	Aka                *string             `gorm:"column:aka" json:"aka"`
	Address            string              `gorm:"column:address;not null" json:"address"`
	PostalCode         string              `gorm:"column:postal_code;not null" json:"postal_code"`
	Phone              *string             `gorm:"column:phone" json:"phone"`
	Website            *string             `gorm:"column:website" json:"website"`
	Streaming          *string             `gorm:"column:streaming" json:"streaming"`
	Notes              *string             `gorm:"column:notes" json:"notes"`
	LivebarnInstalled  bool                `gorm:"column:livebarn_installed" json:"livebarn_installed"`
	LivebarnLocationId int                 `gorm:"column:livebarn_location_id" json:"livebarn_location_id"`
	LivebarnSurfaceId  int                 `gorm:"column:livebarn_surface_id" json:"livebarn_surface_id"`
	HomeTeams          []map[string]string `gorm:"column:home_teams;serializer:json" json:"home_teams"`
	// home teams are array of name and url of home teams.
	Province string `gorm:"column:province" json:"province"`
}

func (MhrLocation) TableName() string {
	return "mhr_locations"
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

type State struct {
	Id    string
	Value string
}

func runScraper(cmd *cobra.Command, args []string) {
	url := BASE_URL + "/rinks"
	resp, err := cl.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		panic(err)
	}

	nodes := htmlquery.Find(doc, `//select[@id="state-select"]/option`)

	wg := sync.WaitGroup{}
	for _, n := range nodes {
		v := htmlutil.GetAttr(n, "value")
		if v == "" {
			continue
		}
		log.Println(v)

		name := htmlquery.InnerText(n)
		state := State{
			Id:    v,
			Value: name,
		}
		wg.Go(func() {
			venues := process_state(state)

			if len(venues) > 0 {
				err := repo.DB.Transaction(func(tx *gorm.DB) error {
					return repo.DB.Save(venues).Error
				})
				if err != nil {
					log.Println("failed to insert to db", err)
				}
			}
		})
	}
	wg.Wait()
	log.Println("done")
}

var reLink = regexp.MustCompile(`rink-info\?r=([0-9]+)`)

func process_state(state State) (result []MhrLocation) {
	url := BASE_URL + "/rinks?state=" + state.Id

	resp, err := cl.Get(url)
	if err != nil {
		log.Printf("url failed %s, %v\n", url, err)
		return nil
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("failed to read body", err)
	}

	doc, err := htmlquery.Parse(bytes.NewReader(b))
	if err != nil {
		log.Printf("failed to parse doc %v\n", err)
	}

	ch := make(chan *MhrLocation, 100)
	wg := sync.WaitGroup{}
	i := 0
	for _, node := range htmlquery.Find(doc, `//main//a`) {
		href := htmlutil.GetAttr(node, "href")

		matches := reLink.FindStringSubmatch(href)
		if matches == nil {
			continue
		}

		i += 1
		wg.Go(func() {
			get_venue(matches[1], ch)
		})
	}
	allVenues := []MhrLocation{}

	for ; i > 0; i -= 1 {
		loc := <-ch
		if loc != nil {
			fmt.Println(loc.RinkName)
			loc.Province = state.Value
			allVenues = append(allVenues, *loc)
		}
	}
	wg.Wait()
	return allVenues
}

func get_venue(mhrId string, ch chan *MhrLocation) {
	var loc *MhrLocation
	var err error
	defer func() {
		ch <- loc
	}()

	url := BASE_URL + "/rink-info?r=" + mhrId
	resp, err := cl.Get(url)
	if err != nil {
		log.Printf("http error %s, %v", url, err)
		return
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("read error %v\n", err)
		return
	}

	doc, err := htmlquery.Parse(bytes.NewReader(b))
	if err != nil {
		log.Printf("failed to parse doc %s , %v\n", url, err)
		return
	}
	loc, err = parse_venue(doc)

	if err != nil {
		log.Printf("failed %s, %v", url, err)
		return
	}
	loc.MhrID, _ = strconv.Atoi(mhrId)
}

var reAddrClean = regexp.MustCompile(`\s+`)

func parse_venue(doc *html.Node) (*MhrLocation, error) {
	loc := &MhrLocation{}

	root := htmlquery.FindOne(doc, `//main/div/div`)
	if root == nil {
		return nil, errors.New("root not found")
	}
	heading := htmlquery.FindOne(root, `/div[2]//h1`)
	if heading == nil {
		return nil, errors.New("heading not found")
	}

	loc.RinkName = htmlquery.InnerText(heading)

	for _, node := range htmlquery.Find(root, `//h3`) {
		label := htmlquery.InnerText(node)

		switch label {
		case "AKA":
			akaNode := htmlquery.FindOne(node, "following-sibling::ul/li")
			if akaNode == nil {
				return nil, errors.New("aka node not found")
			}
			aka := htmlquery.InnerText(akaNode)
			loc.Aka = &aka
		case "Address":
			addrNode := htmlquery.FindOne(node, `following-sibling::p`)
			if addrNode == nil {
				return nil, errors.New("address node not found")
			}
			loc.Address = strings.Trim(htmlquery.InnerText(addrNode), " \n")
			loc.Address = reAddrClean.ReplaceAllString(loc.Address, " ")
			loc.PostalCode = strings.Trim(rePostalCode.FindString(loc.Address), " \n")
		case "Phone":
			phoneNode := htmlquery.FindOne(node, `following-sibling::p`)
			if phoneNode == nil {
				return nil, errors.New("phone node not found")
			}
			phone := strings.Trim(htmlquery.InnerText(phoneNode), "\n ")
			if phone != "" {
				loc.Phone = &phone
			}
		case "Notes":
			notestNode := htmlquery.FindOne(node, `following-sibling::p`)
			if notestNode == nil {
				return nil, errors.New("notes node not found")
			}
			notes := htmlquery.InnerText(notestNode)
			if notes != "" {
				loc.Notes = &notes
			}
		case "Streaming":
			link := htmlquery.FindOne(node, `following-sibling::a`)
			if link == nil {
				log.Println("streaming link not found")
			} else {
				url := htmlutil.GetAttr(link, "href")
				loc.Streaming = &url
				if strings.Contains(url, "livebarn") {
					loc.LivebarnInstalled = true
				}
			}
		case "Home Ice Of":
			ul := htmlquery.FindOne(node, `following-sibling::ul`)
			if ul == nil {
				log.Println("home teams list not found")
			} else {
				var homeTeams = []map[string]string{}
				for _, linkNode := range htmlquery.Find(ul, `//a`) {
					url := htmlutil.GetAttr(linkNode, "href")
					if !strings.Contains(url, "http") {
						url = BASE_URL + "/" + url
					}
					label := htmlquery.InnerText(linkNode)
					homeTeams = append(homeTeams, map[string]string{
						"label": label,
						"url":   url,
					})
				}
				loc.HomeTeams = homeTeams
			}
		}
	}
	websiteNode := htmlquery.FindOne(root, `//a[contains(text(),"Rink Website")]`)
	if websiteNode != nil {
		website := strings.Trim(htmlutil.GetAttr(websiteNode, "href"), "\n ")
		loc.Website = &website
	}

	return loc, nil
}

func runMatcher() {
	var err error

	// Load all locations and unmatched sites_locations into memory for efficient matching
	var allLocations []model.Location
	if err = repo.DB.Where("deleted_at IS NULL").Find(&allLocations).Error; err != nil {
		log.Println(err)
		return
	}

	var unmatchedLocs []MhrLocation
	if err = repo.DB.Where("livebarn_installed=1 AND livebarn_location_id = 0").
		Find(&unmatchedLocs).Error; err != nil {
		log.Println(err)
		return
	}

	type matchResult struct {
		MhrId      int
		locationID int32
	}
	var matches []matchResult

	for _, sl := range unmatchedLocs {
		for i := range allLocations {
			loc := &allLocations[i]
			if sl.PostalCode == "" {
				postalCode := strings.Trim(rePostalCode.FindString(sl.Address), " \n")
				err := repo.DB.Exec(`update mhr_locations set postal_code=? where mhr_id=?`, postalCode, sl.MhrID).Error
				if err != nil {
					log.Fatal(err)
				}
				sl.PostalCode = postalCode
			}
			if sl.PostalCode != "" && strings.Contains(strings.ToLower(loc.PostalCode), strings.ToLower(sl.PostalCode)) {
				matches = append(matches, matchResult{sl.MhrID, loc.ID})
				continue
			}

			if strings.Contains(strings.ToLower(loc.Address1), strings.ToLower(sl.Address)) {
				matches = append(matches, matchResult{sl.MhrID, loc.ID})
				continue
			}

			if strings.Contains(strings.ToLower(loc.Name), strings.ToLower(sl.RinkName)) {
				matches = append(matches, matchResult{sl.MhrID, loc.ID})
				continue
			}
		}
	}

	err = repo.DB.Transaction(func(tx *gorm.DB) error {
		for _, m := range matches {
			if err := tx.Exec("UPDATE mhr_locations SET livebarn_location_id = ? WHERE mhr_id = ?",
				m.locationID, m.MhrId).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}
