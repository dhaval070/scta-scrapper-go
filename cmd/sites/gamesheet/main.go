package main

import (
	"calendar-scrapper/config"
	"calendar-scrapper/dao/model"
	httpclient "calendar-scrapper/internal/client"
	"fmt"
	"log"
	"net/http"
	"os"

	jsoniter "github.com/json-iterator/go"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const URL = "https://gateway.gamesheet.io/stats/schedule?filter[seasons]=%s&filter[start]=2025-11-01&filter[end]=2026-04-30&filter[teams]&filter[divisions]'"

type ScheduleResponse struct {
	Status string `json:"status"`
	Data   []struct {
		Date  string                 `json:"date"`
		Games []jsoniter.RawMessage `json:"games"`
	} `json:"data"`
}

var client = httpclient.GetClient(os.Getenv("HTTP_PROXY"))

func main() {
	// flags := cmdutil.ParseCommonFlags()
	// flag.Parse()

	var err error

	config.Init("config", ".")
	cfg := config.MustReadConfig()

	// Connect to database
	db, err := initDB(&cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Fetch active seasons
	var activeSeasons []model.GamesheetSeason
	if err := db.Where("is_active = ?", 1).Find(&activeSeasons).Error; err != nil {
		log.Fatalf("Failed to fetch active seasons: %v", err)
	}

	log.Printf("Found %d active seasons", len(activeSeasons))

	// Fetch and save schedules for each active season
	for _, season := range activeSeasons {
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
