package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"calendar-scrapper/config"

	"github.com/spf13/cobra"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type claimGame struct {
	GameID         int    `json:"gameId"`
	SurfaceID      string `json:"surfaceId"`
	BroadcastURL   string `json:"broadcastUrl"`
	BroadcasterURL string `json:"broadcasterUrl"`
	VODURL         string `json:"vodUrl"`
	HighlightsURL  string `json:"highlightsUrl"`
}

type claimRequest struct {
	Data struct {
		Games []claimGame `json:"games"`
	} `json:"data"`
}

type ClaimAPILog struct {
	ID             uint      `gorm:"primaryKey;autoIncrement"`
	EventID        string    `gorm:"column:event_id;not null"`
	Site           string    `gorm:"column:site;not null"`
	Datetime       time.Time `gorm:"column:datetime;not null"`
	SurfaceID      int32     `gorm:"column:surface_id;not null"`
	Status         int8      `gorm:"column:status;not null;default:0"`
	HTTPStatusCode *int      `gorm:"column:http_status_code"`
	ResponseBody   *string   `gorm:"column:response_body"`
	ErrorMessage   *string   `gorm:"column:error_message"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (ClaimAPILog) TableName() string {
	return "claim_api_log"
}

type eventRow struct {
	Site      string    `gorm:"column:site"`
	EventID   string    `gorm:"column:event_id"`
	SurfaceID int32     `gorm:"column:surface_id"`
	Datetime  time.Time `gorm:"column:datetime"`
}

var (
	dryRun  bool
	dateStr string
	eventID string
)

var rootCmd = &cobra.Command{
	Use:   "claim-cron",
	Short: "Send games to the GameSheet claim API",
	Long: `Sends eligible games (gs_ sites with surface_id set) to the GameSheet claim API.
Runs as cron by default (batches all today's eligible games), or ad-hoc with --event-id.`,
	Run: runClaimCron,
}

func init() {
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print what would be sent without making the API call")
	rootCmd.Flags().StringVar(&dateStr, "date", "", "Process events for this date (YYYY-MM-DD, default: today)")
	rootCmd.Flags().StringVar(&eventID, "event-id", "", "Process a single event by event_id (adhoc mode)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runClaimCron(cmd *cobra.Command, args []string) {
	config.Init("config", ".")
	cfg := config.MustReadConfig()

	if cfg.GameSheetAPIKey == "" {
		log.Fatal("GAMESHEET_API_KEY is not set in config")
	}

	db, err := gorm.Open(mysql.Open(cfg.DbDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	targetDate := time.Now().Truncate(24 * time.Hour)
	if dateStr != "" {
		parsed, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			log.Fatalf("Invalid --date format (use YYYY-MM-DD): %v", err)
		}
		targetDate = parsed
	}

	var events []eventRow

	if eventID != "" {
		log.Printf("Ad-hoc mode: looking up event_id=%s", eventID)
		var ev eventRow
		result := db.Raw("SELECT site, event_id, surface_id, datetime FROM events WHERE event_id = ?", eventID).Scan(&ev)
		if result.Error != nil {
			log.Fatalf("Database query failed: %v", result.Error)
		}
		if result.RowsAffected == 0 {
			log.Fatalf("No event found with event_id=%s", eventID)
		}
		if !strings.HasPrefix(ev.Site, "gs_") {
			log.Fatalf("Site '%s' does not start with 'gs_' — skipping", ev.Site)
		}
		if ev.SurfaceID == 0 {
			log.Fatalf("Event %s has surface_id=0 — skipping", eventID)
		}
		events = append(events, ev)
	} else {
		log.Printf("Cron mode: querying events for %s", targetDate.Format("2006-01-02"))
		result := db.Raw(`
			SELECT site, event_id, surface_id, datetime
			FROM events
			WHERE site LIKE 'gs_%'
			  AND surface_id IS NOT NULL
			  AND surface_id != 0
			  AND DATE(datetime) = ?
		`, targetDate.Format("2006-01-02")).Scan(&events)
		if result.Error != nil {
			log.Fatalf("Database query failed: %v", result.Error)
		}
	}

	if len(events) == 0 {
		log.Println("No eligible events found")
		return
	}

	log.Printf("Found %d eligible event(s)", len(events))

	games := make([]claimGame, 0, len(events))
	for _, ev := range events {
		games = append(games, buildClaimGame(ev))
	}

	body := claimRequest{}
	body.Data.Games = games

	jsonBody, err := json.Marshal(body)
	if err != nil {
		log.Fatalf("Failed to marshal request body: %v", err)
	}

	apiURL := "https://gateway.gamesheet.io/broadcaster/games"
	log.Printf("--- REQUEST ---")
	log.Printf("POST %s", apiURL)
	log.Printf("Games: %d", len(games))
	log.Printf("Body: %s", string(jsonBody))

	if dryRun {
		log.Println("DRY RUN: skipping API call")
		return
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		log.Fatalf("Failed to create HTTP request: %v", err)
	}
	req.Header.Set("X-GAMESHEET-PARTNER-APIKEY", cfg.GameSheetAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)

	var statusCode int
	var respBodyStr string
	var apiErr error

	if err != nil {
		apiErr = fmt.Errorf("HTTP request failed: %w", err)
		log.Printf("--- ERROR ---")
		log.Printf("%v", apiErr)
		recordResults(db, events, statusCode, respBodyStr, apiErr)
		return
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		apiErr = fmt.Errorf("failed to read response body: %w", err)
		log.Printf("--- ERROR ---")
		log.Printf("%v", apiErr)
		recordResults(db, events, resp.StatusCode, "", apiErr)
		return
	}

	statusCode = resp.StatusCode
	respBodyStr = string(respBytes)

	log.Printf("--- RESPONSE ---")
	log.Printf("Status: %s", resp.Status)
	log.Printf("Status Code: %d", statusCode)
	log.Printf("Body: %s", respBodyStr)

	if statusCode != 200 && statusCode != 201 {
		log.Printf("ERROR: API returned status %d", statusCode)
		recordResults(db, events, statusCode, respBodyStr, fmt.Errorf("API returned status %d", statusCode))
		return
	}

	log.Printf("SUCCESS: %d game(s) claimed successfully", len(events))
	recordResults(db, events, statusCode, respBodyStr, nil)
}

func buildClaimGame(ev eventRow) claimGame {
	gameID, err := strconv.Atoi(ev.EventID)
	if err != nil {
		log.Printf("WARNING: failed to parse event_id '%s' as integer, using 0", ev.EventID)
	}
	broadcastURL := "https://livebarn.com/en/video/" + ev.EventID + "/" + ev.Datetime.Format("2006-01-02") + "/" + ev.Datetime.Format("15:04")
	return claimGame{
		GameID:         gameID,
		SurfaceID:      fmt.Sprintf("%d", ev.SurfaceID),
		BroadcastURL:   broadcastURL,
		BroadcasterURL: "https://livebarn.com/",
		VODURL:         broadcastURL,
		HighlightsURL:  "",
	}
}

func recordResults(db *gorm.DB, events []eventRow, statusCode int, respBody string, apiErr error) {
	status := int8(0)
	httpStatus := statusCode
	var respBodyPtr *string
	if respBody != "" {
		respBodyPtr = &respBody
	}

	if apiErr == nil && statusCode == 200 {
		status = 1
	}

	var errMsg *string
	if apiErr != nil {
		s := apiErr.Error()
		errMsg = &s
	}

	logs := make([]ClaimAPILog, 0, len(events))
	now := time.Now()
	for _, ev := range events {
		logs = append(logs, ClaimAPILog{
			EventID:        ev.EventID,
			Site:           ev.Site,
			Datetime:       ev.Datetime,
			SurfaceID:      ev.SurfaceID,
			Status:         status,
			HTTPStatusCode: &httpStatus,
			ResponseBody:   respBodyPtr,
			ErrorMessage:   errMsg,
			UpdatedAt:      now,
		})
	}

	result := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "event_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"site", "datetime", "surface_id", "status", "http_status_code", "response_body", "error_message", "updated_at"}),
	}).Create(&logs)
	if result.Error != nil {
		log.Printf("ERROR: failed to write to claim_api_log: %v", result.Error)
		return
	}

	plural := "game(s)"
	if len(events) == 1 {
		plural = "game"
	}
	if status == 1 {
		log.Printf("Logged %d %s as success", len(events), plural)
	} else {
		log.Printf("Logged %d %s as failed", len(events), plural)
	}
}
