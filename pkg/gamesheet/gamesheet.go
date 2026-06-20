package gamesheet

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const baseURL = "https://gateway.gamesheet.io/stats/seasons"
const cacheTTL = 5 * time.Minute

var (
	cacheMu  sync.RWMutex
	cached   []Season
	cachedAt time.Time
)

type Season struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	LeagueID      int     `json:"leagueId"`
	IsActive      bool    `json:"is_active"`
	Start         string  `json:"start"`
	End           string  `json:"end"`
	Country       *string `json:"country"`
	StateProvince *string `json:"state_province"`
	AgeCategory   *string `json:"age_category"`
	GameType      *string `json:"game_type"`
}

type seasonsResponse struct {
	Status string   `json:"status"`
	Data   []Season `json:"data"`
}

func fetchFromAPI(apiKey string) ([]Season, error) {
	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("gamesheet: failed to create request: %w", err)
	}
	req.Header.Set("X-Gamesheet-Partner-ApiKey", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gamesheet: failed to fetch seasons: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gamesheet: unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gamesheet: failed to read response: %w", err)
	}

	var result seasonsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("gamesheet: failed to parse response: %w", err)
	}

	return result.Data, nil
}

func FetchSeasons(apiKey string) ([]Season, error) {
	cacheMu.RLock()
	if cached != nil && time.Since(cachedAt) < cacheTTL {
		result := cached
		cacheMu.RUnlock()
		return result, nil
	}
	cacheMu.RUnlock()

	cacheMu.Lock()
	defer cacheMu.Unlock()

	if cached != nil && time.Since(cachedAt) < cacheTTL {
		return cached, nil
	}

	seasons, err := fetchFromAPI(apiKey)
	if err != nil {
		return nil, err
	}

	cached = seasons
	cachedAt = time.Now()
	return seasons, nil
}

func FilterActive(seasons []Season) []Season {
	now := time.Now()
	var active []Season
	for _, s := range seasons {
		if !s.IsActive {
			continue
		}
		endDate, err := time.Parse("2006-01-02", s.End)
		if err != nil || endDate.Before(now) {
			continue
		}
		active = append(active, s)
	}
	return active
}

func FetchActiveSeasons(apiKey string) ([]Season, error) {
	seasons, err := FetchSeasons(apiKey)
	if err != nil {
		return nil, err
	}
	return FilterActive(seasons), nil
}

func FilterByIDs(seasons []Season, excludeIDs []int) []Season {
	if len(excludeIDs) == 0 {
		return seasons
	}
	exclude := make(map[int]struct{}, len(excludeIDs))
	for _, id := range excludeIDs {
		exclude[id] = struct{}{}
	}
	var filtered []Season
	for _, s := range seasons {
		if _, ok := exclude[s.ID]; !ok {
			filtered = append(filtered, s)
		}
	}
	return filtered
}
