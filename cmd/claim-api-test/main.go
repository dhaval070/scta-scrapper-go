package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"calendar-scrapper/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Event struct {
	Site      string    `gorm:"column:site"`
	EventID   string    `gorm:"column:event_id"`
	SurfaceID int32     `gorm:"column:surface_id"`
	HomeTeam  string    `gorm:"column:home_team"`
	GuestTeam string    `gorm:"column:guest_team"`
	Datetime  time.Time `gorm:"column:datetime"`
	Location  string    `gorm:"column:location"`
}

type claimGame struct {
	GameID         int    `json:"gameId"`
	SurfaceID      int    `json:"surfaceId"`
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

func main() {
	eventID := flag.String("event-id", "", "event_id from the events table (required)")
	flag.Parse()

	if *eventID == "" {
		log.Println("ERROR: --event-id is required")
		flag.Usage()
		os.Exit(1)
	}

	log.Printf("Looking up event_id=%s in database...", *eventID)

	config.Init("config", ".")
	cfg := config.MustReadConfig()

	db, err := gorm.Open(mysql.Open(cfg.DbDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("ERROR: failed to connect to database: %v", err)
	}

	var event Event
	result := db.Raw("SELECT site, event_id, surface_id, home_team, guest_team, datetime, location FROM events WHERE event_id = ?", *eventID).Scan(&event)
	if result.Error != nil {
		log.Fatalf("ERROR: database query failed: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		log.Fatalf("ERROR: no event found with event_id=%s", *eventID)
	}

	log.Printf("Found event:")
	log.Printf("  site:       %s", event.Site)
	log.Printf("  event_id:   %s", event.EventID)
	log.Printf("  surface_id: %d", event.SurfaceID)
	log.Printf("  home_team:  %s", event.HomeTeam)
	log.Printf("  guest_team: %s", event.GuestTeam)
	log.Printf("  datetime:   %s", event.Datetime.Format("2006-01-02 15:04:05"))
	log.Printf("  location:   %s", event.Location)

	if !strings.HasPrefix(event.Site, "gs_") {
		log.Fatalf("ERROR: site '%s' does not start with 'gs_' — skipping claim API call", event.Site)
	}

	gameID, err := strconv.Atoi(event.EventID)
	if err != nil {
		log.Fatalf("ERROR: failed to parse event_id '%s' as integer: %v", event.EventID, err)
	}
	// https://livebarn.com/en/video/4424/2026-02-01/10:00
	broadcastUrl := "https://livebarn.com/en/video/" + event.EventID + "/" + event.Datetime.Format("2006-01-02") + "/" + event.Datetime.Format("15:04")

	body := claimRequest{}
	body.Data.Games = []claimGame{
		{
			GameID:         gameID,
			SurfaceID:      int(event.SurfaceID),
			BroadcastURL:   broadcastUrl,
			BroadcasterURL: "https://livebarn.com/",
			VODURL:         broadcastUrl,
			HighlightsURL:  "",
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		log.Fatalf("ERROR: failed to marshal request body: %v", err)
	}

	apiURL := "https://dev-gateway.gamesheet.io/broadcaster/games"
	log.Printf("--- REQUEST ---")
	log.Printf("POST %s", apiURL)
	log.Printf("Body: %s", string(jsonBody))

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		log.Fatalf("ERROR: failed to create HTTP request: %v", err)
	}
	req.Header.Set("X-GAMESHEET-PARTNER-APIKEY", cfg.GameSheetAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("ERROR: HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("ERROR: failed to read response body: %v", err)
	}

	log.Printf("--- RESPONSE ---")
	log.Printf("Status: %s", resp.Status)
	log.Printf("Status Code: %d", resp.StatusCode)
	log.Printf("Headers:")
	for k, v := range resp.Header {
		log.Printf("  %s: %s", k, strings.Join(v, ", "))
	}
	log.Printf("Body: %s", string(respBody))

	if resp.StatusCode >= 400 {
		log.Printf("ERROR: API returned error status %d", resp.StatusCode)
		os.Exit(1)
	}

	log.Printf("SUCCESS: event_id=%s claimed successfully", *eventID)
}
