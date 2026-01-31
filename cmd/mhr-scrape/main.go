package main

import (
	"bytes"
	"calendar-scrapper/config"
	"calendar-scrapper/internal/client"
	"calendar-scrapper/pkg/htmlutil"
	"calendar-scrapper/pkg/repository"
	"errors"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
	"gorm.io/gorm"
)

var cl = client.GetClient("", 5)

const BASE_URL = "https://myhockeyrankings.com"

type MhrLocation struct {
	MhrID             int     `gorm:"column:mhr_id;primaryKey"`
	RinkName          string  `gorm:"column:rink_name;not null"`
	Aka               *string `gorm:"column:aka"`
	Address           string  `gorm:"column:address;not null"`
	Phone             *string `gorm:"column:phone"`
	Website           *string `gorm:"column:website"`
	Streaming         *string `gorm:"column:streaming"`
	Notes             *string `gorm:"column:notes"`
	LivebarnInstalled bool    `gorm:"column:livebarn_installed"`
}

func (MhrLocation) TableName() string {
	return "mhr_locations"
}

func main() {
	config.Init("config", ".")

	var cfg = config.MustReadConfig()
	repo := repository.NewRepository(cfg)

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

	states := make([]string, 0, len(nodes))

	wg := sync.WaitGroup{}
	for _, n := range nodes {
		v := htmlutil.GetAttr(n, "value")
		if v != "" {
			states = append(states, v)
		}
		log.Println(v)
		wg.Go(func() {
			venues := process_state(v)

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
	log.Println("total states: ", len(states))
}

var reLink = regexp.MustCompile(`rink-info\?r=([0-9]+)`)

func process_state(state string) (result []MhrLocation) {
	url := BASE_URL + "/rinks?state=" + state

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
			log.Println(loc.RinkName)
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
	loc.MhrID, _ = strconv.Atoi(mhrId)

	if err != nil {
		log.Printf("failed %s, %v", url, err)
	}
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
			aka := htmlquery.InnerText(node.NextSibling)
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
			notes := htmlquery.InnerText(node.NextSibling)
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
		}
	}
	websiteNode := htmlquery.FindOne(root, `//a[contains(text(),"Rink Website")]`)
	if websiteNode != nil {
		website := strings.Trim(htmlutil.GetAttr(websiteNode, "href"), "\n ")
		loc.Website = &website
	}

	return loc, nil
}
