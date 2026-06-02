package main

import (
	"time"
)

type ClaimStats struct {
	EventsMatched int `json:"events_matched"`
	GamesClaimed  int `json:"games_claimed"`
}

func (app *App) getClaimStats(site string, date time.Time) (*ClaimStats, error) {
	dateStr := date.Format("2006-01-02")

	var eventsMatched int64
	if err := app.db.Raw(
		`SELECT COUNT(*) FROM events WHERE site = ? AND surface_id IS NOT NULL AND surface_id != 0 AND DATE(datetime) = ?`,
		site, dateStr,
	).Scan(&eventsMatched).Error; err != nil {
		return nil, err
	}

	var gamesClaimed int64
	if err := app.db.Raw(
		`SELECT COUNT(*) FROM claim_api_log WHERE status = 1 AND site = ? AND DATE(datetime) = ?`,
		site, dateStr,
	).Scan(&gamesClaimed).Error; err != nil {
		return nil, err
	}

	return &ClaimStats{
		EventsMatched: int(eventsMatched),
		GamesClaimed:  int(gamesClaimed),
	}, nil
}
