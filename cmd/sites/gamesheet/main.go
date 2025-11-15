package main

import (
	"calendar-scrapper/config"
	"calendar-scrapper/dao/model"
	httpclient "calendar-scrapper/internal/client"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const URL = "https://gateway.gamesheet.io/stats/schedule?filter[seasons]=%s&filter[start]=2025-11-01&filter[end]=2026-04-30&filter[teams]&filter[divisions]"

type ScheduleResponse struct {
	Status string `json:"status"`
	Data   []struct {
		Date  string                `json:"date"`
		Games []jsoniter.RawMessage `json:"games"`
	} `json:"data"`
}

var client = httpclient.GetClient("")

func main() {
	seasonsFlag := flag.String("seasons", "", "Season IDs to fetch (use 'all' for all active seasons or comma-separated season IDs like '123,456,789')")
	flag.Parse()

	if *seasonsFlag == "" {
		log.Fatalf("Error: -seasons flag is required. Use 'all' for all active seasons or specify comma-separated season IDs")
	}

	var err error

	config.Init("config", ".")
	cfg := config.MustReadConfig()

	// Connect to database
	db, err := initDB(&cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	var seasonsToProcess []model.GamesheetSeason

	if *seasonsFlag == "all" {
		// Fetch all active seasons
		if err := db.Where("is_active = ?", 1).Find(&seasonsToProcess).Error; err != nil {
			log.Fatalf("Failed to fetch active seasons: %v", err)
		}
		log.Printf("Found %d active seasons", len(seasonsToProcess))
	} else {
		// Parse comma-separated season IDs
		seasonIDStrs := strings.Split(*seasonsFlag, ",")
		for _, idStr := range seasonIDStrs {
			idStr = strings.TrimSpace(idStr)
			seasonID, err := strconv.ParseUint(idStr, 10, 32)
			if err != nil {
				log.Fatalf("Invalid season ID '%s': %v", idStr, err)
			}

			// Fetch specific season
			var season model.GamesheetSeason
			if err := db.First(&season, seasonID).Error; err != nil {
				log.Fatalf("Failed to fetch season %d: %v", seasonID, err)
			}
			seasonsToProcess = append(seasonsToProcess, season)
		}
		log.Printf("Processing %d specific season(s)", len(seasonsToProcess))
	}

	// Fetch and save schedules for each season
	for _, season := range seasonsToProcess {
		log.Printf("Processing Season: ID=%d, Title=%s, LeagueID=%d", season.ID, season.Title, season.LeagueID)

		if err := fetchAndSaveSchedules(db, &cfg, season.ID); err != nil {
			log.Printf("Error fetching schedules for season %d: %v", season.ID, err)
			continue
		}

		log.Printf("Successfully saved schedules for season %d", season.ID)
	}

}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DbDSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func fetchAndSaveSchedules(db *gorm.DB, cfg *config.Config, seasonID uint32) error {
	// Build the API URL
	url := fmt.Sprintf(URL, fmt.Sprintf("%d", seasonID))

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add API key header
	req.Header.Set("X-Gamesheet-Partner-ApiKey", cfg.GameSheetAPIKey)

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch schedules: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse response
	var scheduleResp ScheduleResponse
	if err := jsoniter.NewDecoder(resp.Body).Decode(&scheduleResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Save each game to database
	totalGames := 0
	for _, dayData := range scheduleResp.Data {
		for _, game := range dayData.Games {
			schedule := model.GamesheetSchedule{
				SeasonID: seasonID,
				GameData: model.JSON(game),
			}

			if err := db.Create(&schedule).Error; err != nil {
				log.Printf("Failed to save game: %v", err)
				continue
			}
			totalGames++
		}
	}

	log.Printf("Saved %d games for season %d", totalGames, seasonID)
	return nil
}
