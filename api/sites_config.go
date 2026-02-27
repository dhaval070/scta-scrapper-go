package main

import (
	"database/sql"
	"errors"
	"net/http"
	"surface-api/dao/model"
	"surface-api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// getSitesConfig retrieves all sites config records
func (app *App) getSitesConfig(c *gin.Context) {
	var sitesConfigs []model.SitesConfig

	if err := app.db.Find(&sitesConfigs).Error; err != nil {
		sendError(c, err)
		return
	}

	var response []models.SitesConfigResponse
	for _, sc := range sitesConfigs {
		response = append(response, convertToSitesConfigResponse(sc))
	}

	c.JSON(http.StatusOK, response)
}

// getSitesConfigByID retrieves a single sites config record by ID
func (app *App) getSitesConfigByID(c *gin.Context) {
	id := c.Param("id")
	var sitesConfig model.SitesConfig

	if err := app.db.First(&sitesConfig, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Sites config not found"})
			return
		}
		sendError(c, err)
		return
	}

	c.JSON(http.StatusOK, convertToSitesConfigResponse(sitesConfig))
}

// createSitesConfig creates a new sites config record
func (app *App) createSitesConfig(c *gin.Context) {
	var input models.SitesConfigInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sitesConfig := convertToSitesConfigModel(input)

	if err := app.db.Create(&sitesConfig).Error; err != nil {
		sendError(c, err)
		return
	}

	c.JSON(http.StatusCreated, convertToSitesConfigResponse(sitesConfig))
}

// updateSitesConfig updates an existing sites config record
func (app *App) updateSitesConfig(c *gin.Context) {
	id := c.Param("id")
	var input models.SitesConfigInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var sitesConfig model.SitesConfig
	if err := app.db.First(&sitesConfig, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Sites config not found"})
			return
		}
		sendError(c, err)
		return
	}

	updatedConfig := convertToSitesConfigModel(input)
	updatedConfig.ID = sitesConfig.ID
	updatedConfig.CreatedAt = sitesConfig.CreatedAt

	if err := app.db.Save(&updatedConfig).Error; err != nil {
		sendError(c, err)
		return
	}

	c.JSON(http.StatusOK, convertToSitesConfigResponse(updatedConfig))
}

// deleteSitesConfig deletes a sites config record
func (app *App) deleteSitesConfig(c *gin.Context) {
	id := c.Param("id")
	var sitesConfig model.SitesConfig

	if err := app.db.First(&sitesConfig, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Sites config not found"})
			return
		}
		sendError(c, err)
		return
	}

	if err := app.db.Delete(&sitesConfig).Error; err != nil {
		sendError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sites config deleted successfully"})
}

// getParserTypes returns all available parser types
func (app *App) getParserTypes(c *gin.Context) {
	parserTypes := []string{
		"day_details",
		"day_details_parser1",
		"day_details_parser2",
		"month_based",
		"group_based",
		"custom",
		"external",
	}

	c.JSON(http.StatusOK, parserTypes)
}

// Helper function to convert model to response
func convertToSitesConfigResponse(sc model.SitesConfig) models.SitesConfigResponse {
	response := models.SitesConfigResponse{
		ID:           sc.ID,
		SiteName:     sc.SiteName,
		BaseURL:      sc.BaseURL,
		ParserType:   sc.ParserType,
		CreatedAt:    sc.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    sc.UpdatedAt.Format("2006-01-02 15:04:05"),
		GamesScraped: sc.GamesScraped,
	}

	if sc.DisplayName.Valid {
		response.DisplayName = &sc.DisplayName.String
	}
	if sc.HomeTeam.Valid {
		response.HomeTeam = &sc.HomeTeam.String
	}
	if sc.ParserConfig != nil {
		response.ParserConfig = *sc.ParserConfig
	}
	if sc.Enabled.Valid {
		response.Enabled = &sc.Enabled.Bool
	}
	if sc.LastScrapedAt.Valid {
		lastScraped := sc.LastScrapedAt.Time.Format("2006-01-02 15:04:05")
		response.LastScrapedAt = &lastScraped
	}
	if sc.ScrapeFrequencyHours.Valid {
		response.ScrapeFrequencyHours = &sc.ScrapeFrequencyHours.Int32
	}
	if sc.Notes.Valid {
		response.Notes = &sc.Notes.String
	}

	return response
}

// Helper function to convert input to model
func convertToSitesConfigModel(input models.SitesConfigInput) model.SitesConfig {
	sitesConfig := model.SitesConfig{
		SiteName:   input.SiteName,
		BaseURL:    input.BaseURL,
		ParserType: input.ParserType,
	}

	if input.DisplayName != nil {
		sitesConfig.DisplayName = sql.NullString{String: *input.DisplayName, Valid: true}
	}
	if input.HomeTeam != nil {
		sitesConfig.HomeTeam = sql.NullString{String: *input.HomeTeam, Valid: true}
	}
	if input.ParserConfig != nil {
		pc := model.ParserConfig(input.ParserConfig)
		sitesConfig.ParserConfig = &pc
	}
	if input.Enabled != nil {
		sitesConfig.Enabled = sql.NullBool{Bool: *input.Enabled, Valid: true}
	}
	if input.ScrapeFrequencyHours != nil {
		sitesConfig.ScrapeFrequencyHours = sql.NullInt32{Int32: *input.ScrapeFrequencyHours, Valid: true}
	}
	if input.Notes != nil {
		sitesConfig.Notes = sql.NullString{String: *input.Notes, Valid: true}
	}

	return sitesConfig
}
