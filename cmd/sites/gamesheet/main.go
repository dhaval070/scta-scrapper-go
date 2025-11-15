package main

import (
	"calendar-scrapper/config"
	"calendar-scrapper/dao/model"
	httpclient "calendar-scrapper/internal/client"
	"calendar-scrapper/pkg/cmdutil"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

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

type Game struct {
	ID                 int64     `json:"id"`
	GameType           string    `json:"gameType"`
	Number             string    `json:"number"`
	Location           string    `json:"location"`
	ScheduledStartTime time.Time `json:"scheduledStartTime"`
	StartTime          time.Time `json:"startTime"`
	EndTime            time.Time `json:"endTime"`
	Status             string    `json:"status"`
	Category           string    `json:"category"`
	HasOvertime        bool      `json:"hasOvertime"`
	HasShootout        bool      `json:"hasShootout"`
	Periods            struct {
		Period1 int `json:"1"`
		Period2 int `json:"2"`
		Period3 int `json:"3"`
		OT1     int `json:"ot_1"`
	} `json:"periods"`
	Home    Team `json:"home"`
	Visitor Team `json:"visitor"`
}

type Team struct {
	ID       int64    `json:"id"`
	Title    string   `json:"title"`
	Logo     string   `json:"logo"`
	Division Division `json:"division"`
	Record   Record   `json:"record"`
}

type Division struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

type Record struct {
	GamesPlayed            int `json:"gamesPlayed"`
	Wins                   int `json:"wins"`
	Losses                 int `json:"losses"`
	Ties                   int `json:"ties"`
	OvertimeShootoutWins   int `json:"overtimeShootoutWins"`
	OvertimeShootoutLosses int `json:"overtimeShootoutLosses"`
}

var client = httpclient.GetClient("")

func main() {
	seasonsFlag := flag.String("seasons", "", "Season IDs to fetch (use 'all' for all active seasons or comma-separated season IDs like '123,456,789')")
	outfileFlag := flag.String("outfile", "", "Output CSV file path(optional)")
	importLocationsFlag := flag.Bool("import-locations", false, "Import locations to database")
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

		result, err := fetchAndSaveSchedules(db, &cfg, season, *importLocationsFlag)
		if err != nil {
			log.Printf("Error fetching schedules for season %d: %v", season.ID, err)
			continue
		}

		if *importLocationsFlag {
			if err := cmdutil.ImportLocations(season.Site, result); err != nil {
				log.Fatal(err)
			}
		}
		var filename string

		if *outfileFlag == "-" {
			filename = "-"
		} else {
			filename = season.Site + "_" + *outfileFlag
		}

		if err := cmdutil.WriteOutput(filename, result); err != nil {
			log.Fatal(err)
		}

		log.Printf("Successfully processed schedules for season %d", season.ID)
	}

}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DbDSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func fetchAndSaveSchedules(db *gorm.DB, cfg *config.Config, season model.GamesheetSeason, importLocations bool) ([][]string, error) {
	seasonID := season.ID
	// Build the API URL
	url := fmt.Sprintf(URL, fmt.Sprintf("%d", seasonID))

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add API key header
	req.Header.Set("X-Gamesheet-Partner-ApiKey", cfg.GameSheetAPIKey)

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schedules: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse response
	var scheduleResp ScheduleResponse
	if err := jsoniter.NewDecoder(resp.Body).Decode(&scheduleResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Process each game
	totalGames := 0

	var result = [][]string{}

	for _, dayData := range scheduleResp.Data {
		for _, gameRaw := range dayData.Games {
			var game Game
			if err := jsoniter.Unmarshal(gameRaw, &game); err != nil {
				log.Printf("Failed to decode game: %v", err)
				continue
			}

			if game.Location == "" {
				log.Println("empty location found: ", string(gameRaw))
			}

			// Save to database
			schedule := model.GamesheetSchedule{
				SeasonID: seasonID,
				GameData: model.JSON(gameRaw),
			}

			if err := db.Create(&schedule).Error; err != nil {
				log.Printf("Failed to save game: %v", err)
				continue
			}

			result = append(result, []string{game.ScheduledStartTime.Format("2006-01-02 15:04"), season.Site, game.Home.Title, game.Visitor.Title, game.Location, game.Home.Division.Title, ""})
			totalGames++
		}
	}

	log.Printf("Saved %d games for season %d", totalGames, seasonID)

	return result, nil
}
