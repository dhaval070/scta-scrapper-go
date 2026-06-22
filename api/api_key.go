package main

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strconv"
	"surface-api/dao/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// APIKeyMiddleware validates the X-API-Key header against the api_keys table
func (app *App) APIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing API key"})
			return
		}

		hash := sha256.Sum256([]byte(apiKey))
		hashStr := fmt.Sprintf("%x", hash)

		var key model.ApiKey
		if err := app.db.Where("key_hash = ? AND active = 1", hashStr).First(&key).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			return
		}

		c.Next()
	}
}

// exportKmasterVenuesAPI returns kmaster venues with LiveBarn location and surface details via API key auth
func (app *App) exportKmasterVenuesAPI(c *gin.Context) {
	country := c.Query("country")
	state := c.Query("state")
	livebarn := c.Query("livebarn")

	query := app.db.Model(&model.KmasterVenueList{})
	if country != "" {
		query = query.Where("country = ?", country)
	}
	if state != "" {
		query = query.Where("province_state = ?", state)
	}
	if livebarn != "" {
		livebarnBool, err := strconv.ParseBool(livebarn)
		if err == nil {
			if livebarnBool {
				query = query.Where("livebarn_venue_id IN (?) AND livebarn_venue_id != 0",
					app.db.Model(&model.Location{}).Select("id"))
			} else {
				query = query.Where("livebarn_venue_id NOT IN (?) OR livebarn_venue_id = 0",
					app.db.Model(&model.Location{}).Select("id"))
			}
		}
	}

	var total int64
	query.Count(&total)

	var venues []model.KmasterVenueList
	if err := query.Session(&gorm.Session{}).Order("id DESC").Find(&venues).Error; err != nil {
		sendError(c, err)
		return
	}

	result := app.buildKmasterVenueExport(venues, total)
	c.JSON(http.StatusOK, result)
}
