package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"calendar-scrapper/pkg/gamesheet"
	"surface-api/dao/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Season represents a gamesheet season used in swagger docs.
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

// getGamesheetSeasons returns active gamesheet seasons
// @Summary Get active gamesheet seasons
// @Description Fetches all seasons from the Gamesheet API and filters out inactive or expired seasons
// @Tags Gamesheet
// @Accept json
// @Produce json
// @Param exclude_existing query bool false "Exclude seasons already stored in the local gamesheet_seasons table"
// @Success 200 {array} Season
// @Failure 500 {object} map[string]string
// @Security CookieAuth
// @Router /gamesheet-seasons [get]
func (app *App) getGamesheetSeasons(c *gin.Context) {
	seasons, err := gamesheet.FetchActiveSeasons(app.cfg.GameSheetAPIKey)
	if err != nil {
		sendError(c, err)
		return
	}

	if c.Query("exclude_existing") == "true" {
		var existing []model.GamesheetSeason
		app.db.Select("id").Find(&existing)
		ids := make([]int, len(existing))
		for i, s := range existing {
			ids[i] = int(s.ID)
		}
		seasons = gamesheet.FilterByIDs(seasons, ids)
	}

	c.JSON(http.StatusOK, seasons)
}

// importGamesheetSeasons imports seasons into gamesheet_seasons and sites_config
// @Summary Import gamesheet seasons
// @Description Accepts an array of Season objects and inserts them into gamesheet_seasons and sites_config tables within a single transaction. Skips records that already exist.
// @Tags Gamesheet
// @Accept json
// @Produce json
// @Param body body []Season true "Array of seasons to import"
// @Success 201 {object} map[string]int "Number of seasons imported"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security CookieAuth
// @Router /gamesheet-seasons/import [post]
func (app *App) importGamesheetSeasons(c *gin.Context) {
	var seasons []Season
	if err := c.BindJSON(&seasons); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(seasons) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty seasons array"})
		return
	}

	var existing []model.GamesheetSeason
	app.db.Select("id").Find(&existing)
	existingIDs := make(map[int32]struct{}, len(existing))
	for _, e := range existing {
		existingIDs[e.ID] = struct{}{}
	}

	imported := 0

	err := app.db.Transaction(func(tx *gorm.DB) error {
		for _, s := range seasons {
			if _, exists := existingIDs[int32(s.ID)]; exists {
				continue
			}

			site := fmt.Sprintf("gs_%d", s.ID)

			startDate, _ := time.Parse("2006-01-02", s.Start)
			endDate, _ := time.Parse("2006-01-02", s.End)

			isActive := int32(0)
			if s.IsActive {
				isActive = 1
			}

			gs := model.GamesheetSeason{
				ID:        int32(s.ID),
				Title:     s.Title,
				Site:      site,
				LeagueID:  int32(s.LeagueID),
				IsActive:  isActive,
				StartDate: startDate,
				EndDate:   endDate,
			}

			if err := tx.Create(&gs).Error; err != nil {
				return err
			}

			sc := model.SitesConfig{
				SiteName:    site,
				DisplayName: sql.NullString{String: s.Title, Valid: true},
				BaseURL:     "",
				ParserType:  "external",
				ParserConfig: &model.ParserConfig{
					"season_id":   s.ID,
					"binary_path": "./bin/gamesheet",
					"extra_args":  []string{fmt.Sprintf("--sites=%s", site)},
				},
				Enabled:              sql.NullBool{Bool: s.IsActive, Valid: true},
				ScrapeFrequencyHours: sql.NullInt32{Int32: 24, Valid: true},
			}

			if err := tx.Create(&sc).Error; err != nil {
				return err
			}

			imported++
		}
		return nil
	})

	if err != nil {
		sendError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"imported": imported})
}
