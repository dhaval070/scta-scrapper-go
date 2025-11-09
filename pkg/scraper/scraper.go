package scraper

import (
	"bytes"
	httpclient "calendar-scrapper/internal/client"
	"calendar-scrapper/pkg/parser"
	"calendar-scrapper/pkg/parser1"
	"calendar-scrapper/pkg/parser2"
	"calendar-scrapper/pkg/siteconfig"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
)

var client = httpclient.GetClient(os.Getenv("HTTP_PROXY"))

// Scraper handles dynamic scraping based on site configuration
type Scraper struct {
	config    *siteconfig.SiteConfig
	parserCfg *siteconfig.ParserConfigJSON
	loader    *siteconfig.Loader
}

// New creates a new scraper instance for a site
func New(config *siteconfig.SiteConfig, loader *siteconfig.Loader) (*Scraper, error) {
	parserCfg, err := loader.GetParserConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config for %s: %w", config.SiteName, err)
	}

	return &Scraper{
		config:    config,
		parserCfg: parserCfg,
		loader:    loader,
	}, nil
}

// Scrape performs the scraping based on the site's parser type
func (s *Scraper) Scrape(mm, yyyy int) ([][]string, error) {
	log.Printf("[%s] Starting scrape for %d/%d\n", s.config.SiteName, mm, yyyy)

	switch s.config.ParserType {
	case siteconfig.ParserTypeDayDetails:
		return s.scrapeDayDetails(mm, yyyy)
	case siteconfig.ParserTypeDayDetailsParser1:
		return s.scrapeDayDetailsParser1(mm, yyyy)
	case siteconfig.ParserTypeDayDetailsParser2:
		return s.scrapeDayDetailsParser2(mm, yyyy)
	case siteconfig.ParserTypeGroupBased:
		return s.scrapeGroupBased(mm, yyyy)
	case siteconfig.ParserTypeMonthBased:
		return s.scrapeMonthBased(mm, yyyy)
	case siteconfig.ParserTypeExternal:
		return s.scrapeExternal(mm, yyyy)
	default:
		return nil, fmt.Errorf("unsupported parser type: %s", s.config.ParserType)
	}
}

func loadUrl(url string) (io.Reader, error) {
	var try int
	var resp *http.Response
	var err error
	for try = 1; try < 4; try += 1 {
		resp, err = client.Get(url)
		if err != nil {
			log.Printf("error: load url failed: url=%s, err=%v\n", url, err)
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}
	if err != nil {
		log.Printf("error: loadurl - all retry failed url=%s, err=%v\n", url, err)
		return nil, err
	}
	if try > 1 {
		log.Printf("retry successful: url=%v\n", url)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("error: failed to read body, url=%s, err=%v\n", url, err)
	}
	return bytes.NewReader(body), err
}

// scrapeDayDetails handles day_details parser type
func (s *Scraper) scrapeDayDetails(mm, yyyy int) ([][]string, error) {
	url := fmt.Sprintf(s.config.BaseURL+s.parserCfg.URLTemplate, mm, yyyy)

	log.Printf("[%s] Loading URL: %s\n", s.config.SiteName, url)
	body, err := loadUrl(url)
	if err != nil {
		return nil, err
	}
	doc, err := htmlquery.Parse(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	cfg := parser.DayDetailsConfig{
		TournamentCheckExact: s.parserCfg.TournamentCheckExact,
		LogErrors:            s.parserCfg.LogErrors,
		ContentFilter:        s.parserCfg.ContentFilter,
		// GameDetailsFunc: func(gameURL string) string {
		// 	return parser.GetGameDetailsAddress(gameURL, s.config.BaseURL)
		// },
		VenueAddressFunc: func(url, class string) string {
			if len(url) > 3 && url[:4] != "http" {
				url = s.config.BaseURL + url
			}
			address, err := parser.VenueFetcher.Fetch(url, class)
			if err != nil {
				return ""
			}
			return address
		},
	}

	result := parser.ParseDayDetailsSchedule(
		doc,
		s.config.SiteName,
		s.config.BaseURL,
		s.config.HomeTeam,
		cfg,
	)

	log.Printf("[%s] Parsed %d schedule entries\n", s.config.SiteName, len(result))
	return result, nil
}

// scrapeGroupBased handles group_based parser type
func (s *Scraper) scrapeGroupBased(mm, yyyy int) ([][]string, error) {
	seasonsURL := s.config.BaseURL + s.parserCfg.SeasonsURL

	log.Printf("[%s] Loading seasons URL: %s\n", s.config.SiteName, seasonsURL)
	body, err := loadUrl(seasonsURL)
	if err != nil {
		return nil, err
	}
	doc, err := htmlquery.Parse(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	log.Printf("[%s] Parsing groups with XPath: %s\n", s.config.SiteName, s.parserCfg.GroupXPath)
	groups := parser.ParseSiteListGroups(doc, s.parserCfg.GroupXPath)

	if len(groups) == 0 {
		return nil, fmt.Errorf("no groups found")
	}

	log.Printf("[%s] Found %d groups\n", s.config.SiteName, len(groups))

	urlTemplate := s.config.BaseURL + s.parserCfg.GroupURLTemplate
	result := parser.FetchSchedules(s.config.SiteName, s.config.BaseURL, urlTemplate, groups, mm, yyyy)

	log.Printf("[%s] Parsed %d schedule entries from all groups\n", s.config.SiteName, len(result))
	return result, nil
}

// scrapeMonthBased handles month_based parser type
func (s *Scraper) scrapeMonthBased(mm, yyyy int) ([][]string, error) {
	url := fmt.Sprintf(s.config.BaseURL+s.parserCfg.URLTemplate, mm, yyyy)

	log.Printf("[%s] Loading URL: %s\n", s.config.SiteName, url)
	body, err := loadUrl(url)
	if err != nil {
		return nil, err
	}
	doc, err := htmlquery.Parse(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	cfg := parser.MonthScheduleConfig{
		TeamParseStrategy: s.parserCfg.TeamParseStrategy,
		VenueAddressFunc: func(url, class string) string {
			if len(url) > 3 && url[:4] != "http" {
				url = s.config.BaseURL + url
			}
			address, err := parser.VenueFetcher.Fetch(url, class)
			if err != nil {
				return ""
			}
			return address
		},
	}

	result := parser.ParseMonthBasedSchedule(doc, mm, yyyy, s.config.SiteName, cfg)

	log.Printf("[%s] Parsed %d schedule entries\n", s.config.SiteName, len(result))
	return result, nil
}

// scrapeDayDetailsParser1 handles day_details_parser1 parser type (uses parser1 package)
func (s *Scraper) scrapeDayDetailsParser1(mm, yyyy int) ([][]string, error) {
	url := fmt.Sprintf(s.config.BaseURL+s.parserCfg.URLTemplate, mm, yyyy)

	log.Printf("[%s] Loading URL: %s\n", s.config.SiteName, url)
	body, err := loadUrl(url)
	if err != nil {
		return nil, err
	}
	doc, err := htmlquery.Parse(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// parser1.ParseSchedules has different signature - includes baseURL and homeTeam
	result := parser1.ParseSchedules(doc, s.config.SiteName, s.config.BaseURL, s.config.HomeTeam)

	log.Printf("[%s] Parsed %d schedule entries\n", s.config.SiteName, len(result))
	return result, nil
}

// scrapeDayDetailsParser2 handles day_details_parser2 parser type (uses parser2 package)
// This parser requires explicit "home game" or "away game" markers
func (s *Scraper) scrapeDayDetailsParser2(mm, yyyy int) ([][]string, error) {
	url := fmt.Sprintf(s.config.BaseURL+s.parserCfg.URLTemplate, mm, yyyy)

	log.Printf("[%s] Loading URL: %s\n", s.config.SiteName, url)

	body, err := loadUrl(url)
	if err != nil {
		return nil, err
	}
	doc, err := htmlquery.Parse(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// parser2.ParseSchedules has same signature as parser1
	result := parser2.ParseSchedules(doc, s.config.SiteName, s.config.BaseURL)

	log.Printf("[%s] Parsed %d schedule entries\n", s.config.SiteName, len(result))
	return result, nil
}

// scrapeExternal calls an external binary to scrape the site
func (s *Scraper) scrapeExternal(mm, yyyy int) ([][]string, error) {
	if s.parserCfg.BinaryPath == "" {
		return nil, fmt.Errorf("binary_path not configured for external parser")
	}

	// Build command arguments
	dateStr := fmt.Sprintf("%02d%d", mm, yyyy)
	args := []string{"--date", dateStr, "--outfile", "-"}
	args = append(args, s.parserCfg.ExtraArgs...)

	log.Printf("[%s] Calling external binary: %s %v\n", s.config.SiteName, s.parserCfg.BinaryPath, args)

	// Execute the binary
	cmd := exec.Command(s.parserCfg.BinaryPath, args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("external binary failed: %s\nstderr: %s", err, string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to execute external binary: %w", err)
	}

	// Parse CSV output
	reader := csv.NewReader(strings.NewReader(string(output)))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV output: %w", err)
	}

	log.Printf("[%s] External parser returned %d records\n", s.config.SiteName, len(records))
	return records, nil
}

// GetConfig returns the site configuration
func (s *Scraper) GetConfig() *siteconfig.SiteConfig {
	return s.config
}
