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

var (
	matchAddress bool
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
		runMatcherGamesheetMethod(matchAddress)
		// if gamesheet {
		// 	log.Println("matching by gamesheet method")
		// 	runMatcherGamesheetMethod(gamesheet)
		// } else {
		// 	log.Println("matching by address method")
		// 	runMatcherAddressMethod()
		// }
	},
}

var cfg config.Config
var repo *repository.Repository

const BASE_URL = "https://myhockeyrankings.com"

var cl = client.GetClient("", 5)

func init() {
	matchCmd.Flags().BoolVar(&matchAddress, "address", false, "force address match along with rink name")
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

//
// func runMatcherAddressMethod() {
// 	var err error
// 	// Load all locations and unmatched sites_locations into memory for efficient matching
// 	var allLocations []model.Location
// 	if err = repo.DB.Where("deleted_at IS NULL").Find(&allLocations).Error; err != nil {
// 		log.Println(err)
// 		return
// 	}
//
// 	var unmatchedLocs []MhrLocation
// 	if err = repo.DB.Where("livebarn_location_id = 0").Find(&unmatchedLocs).Error; err != nil {
// 		log.Println(err)
// 	}
//
// 	// Match in memory - find best (longest) location name match for each site location
// 	type matchResult struct {
// 		Location   string
// 		locationID int32
// 	}
// 	var matches []matchResult
//
// 	rePostalCode := regexp.MustCompile(`([0-9]{5,})\s*$`)
//
// 	for _, sl := range unmatchedLocs {
// 		var bestMatch *model.Location
// 		var bestLen int
// 		for i := range allLocations {
// 			loc := &allLocations[i]
// 			matches := rePostalCode.FindStringSubmatch(sl.Address)
//
// 			if len(matches) < 2 {
// 				continue
// 			}
//
// 			if strings.Contains(strings.ToLower(sl.RinkName), strings.ToLower(loc.Name)) && strings.Contains(loc.PostalCode, matches[1]) {
// 				if len(loc.Name) > bestLen {
// 					bestMatch = loc
// 					bestLen = len(loc.Name)
// 				}
// 			}
// 		}
// 		if bestMatch != nil {
// 			matches = append(matches, matchResult{sl.RinkName, bestMatch.ID})
// 		}
// 	}
//
// }

func runMatcherGamesheetMethod(matchAddress bool) {
	var err error

	// Load all locations and unmatched sites_locations into memory for efficient matching
	var allLocations []model.Location
	if err = repo.DB.Where("deleted_at IS NULL").Find(&allLocations).Error; err != nil {
		log.Println(err)
		return
	}

	var unmatchedLocs []MhrLocation
	if err = repo.DB.Where("livebarn_location_id = 0").Find(&unmatchedLocs).Error; err != nil {
		log.Println(err)
	}

	// Match in memory - find best (longest) location name match for each site location
	type matchResult struct {
		Location   string
		locationID int32
	}
	var matches []matchResult
	rePostalCode := regexp.MustCompile(`([0-9]{5,})\s*$`)

	for _, sl := range unmatchedLocs {
		var bestMatch *model.Location
		var bestLen int
		for i := range allLocations {
			loc := &allLocations[i]
			// if gamesheet method then don't match postal code.
			if matchAddress {
				matches := rePostalCode.FindStringSubmatch(sl.Address)

				if len(matches) < 2 {
					continue
				}

				if strings.Contains(strings.ToLower(sl.RinkName), strings.ToLower(loc.Name)) &&
					strings.Contains(loc.PostalCode, matches[1]) {
					if len(loc.Name) > bestLen {
						bestMatch = loc
						bestLen = len(loc.Name)
					}
				}
			} else {
				if strings.Contains(strings.ToLower(sl.RinkName), strings.ToLower(loc.Name)) {
					if len(loc.Name) > bestLen {
						bestMatch = loc
						bestLen = len(loc.Name)
					}
				}
			}
		}
		if bestMatch != nil {
			matches = append(matches, matchResult{sl.RinkName, bestMatch.ID})
		}
	}

	// Batch update matched location_ids
	err = repo.DB.Transaction(func(tx *gorm.DB) error {
		for _, m := range matches {
			if err := tx.Exec("UPDATE mhr_locations SET livebarn_location_id = ? WHERE rink_name = ? AND livebarn_location_id != -1",
				m.locationID, m.Location).Error; err != nil {
				return err
			}
		}

		// set surface id if matched location has just 1 surface.
		err = tx.Exec(`UPDATE mhr_locations sl
			JOIN surfaces s ON sl.livebarn_location_id = s.location_id AND s.deleted_at IS NULL
			JOIN (SELECT location_id FROM surfaces WHERE deleted_at IS NULL GROUP BY location_id HAVING COUNT(*) = 1) single 
				ON sl.livebarn_location_id = single.location_id
			SET sl.livebarn_surface_id = s.id
			WHERE sl.livebarn_location_id > 0`).Error
		if err != nil {
			return err
		}

		// set surface id where remaining part of surface location matches with surface name
		err = tx.Exec(`UPDATE mhr_locations sl, locations l, surfaces s
			SET sl.livebarn_surface_id=s.id
			WHERE
			sl.livebarn_location_id=l.id AND s.location_id=l.id
			AND sl.livebarn_surface_id=0 AND sl.livebarn_location_id > 0
			AND s.deleted_at IS NULL
			AND locate(s.name, trim(replace(sl.rink_name, l.name, '')))>0`).Error
		return err
	})
	if err != nil {
		log.Println(err)
		return
	}

	var siteLoc []MhrLocation
	err = repo.DB.Raw(`SELECT rink_name, livebarn_location_id FROM mhr_locations WHERE
		livebarn_surface_id=0 AND livebarn_location_id > 0`).Scan(&siteLoc).Error
	if err != nil {
		log.Println(err)
		return
	}

	var ids = make([]int, 0, len(allLocations))
	for _, l := range allLocations {
		ids = append(ids, int(l.ID))
	}

	var surfaces = []model.Surface{}

	err = repo.DB.Where("location_id in ? AND deleted_at IS NULL", ids).Find(&surfaces).Error

	smap := map[int32][]model.Surface{}

	for _, s := range surfaces {
		smap[s.LocationID] = append(smap[s.LocationID], s)
	}

	// Create location map for fast lookup
	locMap := map[int32]*model.Location{}
	for i := range allLocations {
		locMap[allLocations[i].ID] = &allLocations[i]
	}

	// var totalLocMatch, totalSurfaceMatch = 0, 0
	for _, sl := range siteLoc {
		id := int32(sl.LivebarnLocationId)

		words := strings.Split(sl.RinkName, " ")
		if len(words) == 0 {
			continue
		}

		err := repo.DB.Transaction(func(tx *gorm.DB) error {
			_, err = setSurface(sl, id, smap, locMap, words[len(words)-1], tx)
			return err
		})

		if err != nil {
			log.Println(err)
			return
		}
		// for MHR we are not currently not matching location by tokens. uncomment following code to match by tokens

		// locMatched, surfaceMatched, err := MatchLocByTokens(sl, allLocations, smap, locMap)
		// if err != nil {
		// 	log.Println(err)
		// 	return
		// }
		//
		// if locMatched {
		// 	totalLocMatch++
		// }
		// if surfaceMatched {
		// 	totalSurfaceMatch++
		// }
	}
}

var blackList = map[string]bool{
	"ice": true, "arena": true, "pavilion": true, "centennial": true,
	"arctic": true, "national": true, "sports": true, "sportplex": true,
	"sportsplex": true,
	"bell":       true, "center": true, "centre": true, "field": true,
	"fields": true, "livebarn": true, "convention": true,
}

// splits slite location words and match with each livebarn location. also sets surface id by matching last word in site location.
func MatchLocByTokens(sl MhrLocation, locations []model.Location, smap map[int32][]model.Surface, locMap map[int32]*model.Location) (bool, bool, error) {
	var err error

	tokens := strings.Split(sl.RinkName, " ")
	lastWord := tokens[len(tokens)-1]

	locMatched := false
	surfaceMatched := false

TOKENS_LOOP:
	// match each token with livebarn location.
	for _, t := range tokens {
		if len(t) == 0 || reNonAlphaNum.MatchString(t) || blackList[strings.ToLower(t)] {
			continue
		}
		var id int32 = 0

		for _, l := range locations {
			if strings.Contains(l.Name, t) {
				// if more than one location matched then skip token.
				if id != 0 {
					continue TOKENS_LOOP
				}
				id = l.ID
			}
		}
		if id == 0 {
			continue
		}

		err = repo.DB.Transaction(func(tx *gorm.DB) error {
			// set location id
			err = tx.Exec(`UPDATE mhr_locations set livebarn_location_id=? WHERE rink_name=? AND livebarn_location_id != -1`, id, sl.RinkName).Error
			if err != nil {
				return fmt.Errorf("failed to set location id, %w", err)
			}
			locMatched = true

			surfaceMatched, err = setSurface(sl, id, smap, locMap, lastWord, tx)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return false, false, err
		}
		break
	}
	return locMatched, surfaceMatched, nil
}

func setSurface(sl MhrLocation, locId int32, smap map[int32][]model.Surface, locMap map[int32]*model.Location, lastWord string, tx *gorm.DB) (bool, error) {
	if len(smap[locId]) == 1 {
		return setSingleSurface(sl, smap[locId][0].ID, tx)
	}

	if matched, err := matchBySanitizedName(sl, locId, smap, locMap, tx); err != nil || matched {
		return matched, err
	}

	return matchByLastWord(sl, locId, smap, lastWord, tx)
}

func setSingleSurface(sl MhrLocation, surfaceID int32, tx *gorm.DB) (bool, error) {
	err := tx.Exec(`UPDATE mhr_locations SET livebarn_surface_id=? WHERE rink_name=? AND livebarn_surface_id != -1`,
		surfaceID, sl.RinkName).Error
	if err != nil {
		return false, fmt.Errorf("failed to set location id, %w", err)
	}
	return true, nil
}

var reNonAlphaNum = regexp.MustCompile("[^0-9A-Za-z]")

func matchBySanitizedName(sl MhrLocation, locId int32, smap map[int32][]model.Surface, locMap map[int32]*model.Location, tx *gorm.DB) (bool, error) {
	location, ok := locMap[locId]
	if !ok {
		return false, nil
	}

	remainingPart := strings.TrimSpace(strings.ReplaceAll(sl.RinkName, location.Name, ""))
	sanitizedLocName := strings.ToLower(reNonAlphaNum.ReplaceAllString(remainingPart, ""))

	if sanitizedLocName == "" {
		return false, nil
	}

	for _, s := range smap[locId] {
		sanitizedSurfaceName := strings.ToLower(reNonAlphaNum.ReplaceAllString(s.Name, ""))

		if sanitizedSurfaceName != "" && strings.Contains(sanitizedLocName, sanitizedSurfaceName) {
			err := tx.Exec(`UPDATE mhr_locations SET livebarn_surface_id=? WHERE rink_name=? AND livebarn_surface_id=0`,
				s.ID, sl.RinkName).Error
			if err != nil {
				return false, fmt.Errorf("failed to set surface id, %w", err)
			}
			return true, nil
		}
	}
	return false, nil
}

func matchByLastWord(sl MhrLocation, locId int32, smap map[int32][]model.Surface, lastWord string, tx *gorm.DB) (bool, error) {
	lastWord = reNonAlphaNum.ReplaceAllString(lastWord, "")
	if lastWord == "" {
		return false, nil
	}

	for _, s := range smap[locId] {
		if !strings.Contains(strings.ToLower(s.Name), strings.ToLower(lastWord)) {
			continue
		}

		err := tx.Exec(`UPDATE mhr_locations SET livebarn_surface_id=? WHERE rink_name=? AND livebarn_surface_id=0`,
			s.ID, sl.RinkName).Error
		if err != nil {
			return false, fmt.Errorf("failed to set surface id, %w", err)
		}
		return true, nil
	}
	return false, nil
}
