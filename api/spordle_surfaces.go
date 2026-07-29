package main

import (
	"net/http"
	"strconv"
	"surface-api/models"

	"surface-api/dao/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// getSpordleSurfaces retrieves paginated spordle surface records
// @Summary List spordle surfaces
// @Description Returns a paginated list of surfaces from the spordle surfaces table
// @Tags SpordleSurface
// @Accept json
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param perPage query int false "Items per page (default 10, max 100)"
// @Param name query string false "Filter by venue name (partial match)"
// @Param city query string false "Filter by venue city (exact match)"
// @Param region query string false "Filter by venue region/province (exact match)"
// @Param country query string false "Filter by venue country (exact match)"
// @Param sport query string false "Filter by surface sport (partial match)"
// @Param surfaceType query string false "Filter by surface type (exact match, e.g. Ice, Grass, Synthetic)"
// @Param surfaceSize query string false "Filter by surface size (exact match, e.g. XS, S, M, L, XL)"
// @Success 200 {object} models.SpordleSurfaceResult "Paginated spordle surface list"
// @Failure 500 {object} object "Internal server error"
// @Security CookieAuth
// @Router /spordle-surfaces [get]
func (app *App) getSpordleSurfaces(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("perPage", "10")

	pageNum, _ := strconv.Atoi(page)
	perPageNum, _ := strconv.Atoi(perPage)
	if pageNum < 1 {
		pageNum = 1
	}
	if perPageNum < 1 || perPageNum > 100 {
		perPageNum = 10
	}
	offset := (pageNum - 1) * perPageNum

	name := c.Query("name")
	city := c.Query("city")
	region := c.Query("region")
	country := c.Query("country")
	sport := c.Query("sport")
	surfaceType := c.Query("surfaceType")
	surfaceSize := c.Query("surfaceSize")

	query := app.db.Model(&model.SpordleSurface{})
	if name != "" {
		query = query.Where("venue_name LIKE ?", "%"+name+"%")
	}
	if city != "" {
		query = query.Where("venue_city = ?", city)
	}
	if region != "" {
		query = query.Where("venue_region = ?", region)
	}
	if country != "" {
		query = query.Where("venue_country = ?", country)
	}
	if sport != "" {
		query = query.Where("surface_sports LIKE ?", "%"+sport+"%")
	}
	if surfaceType != "" {
		query = query.Where("surface_type = ?", surfaceType)
	}
	if surfaceSize != "" {
		query = query.Where("surface_size = ?", surfaceSize)
	}

	var total int64
	query.Count(&total)

	var surfaces []model.SpordleSurface
	if err := query.Session(&gorm.Session{}).Offset(offset).Limit(perPageNum).Order("venue_name ASC").Find(&surfaces).Error; err != nil {
		sendError(c, err)
		return
	}

	var response []models.SpordleSurfaceResponse
	for _, s := range surfaces {
		response = append(response, convertToSpordleSurfaceResponse(s))
	}

	c.JSON(http.StatusOK, models.SpordleSurfaceResult{
		Data:    response,
		Page:    pageNum,
		PerPage: perPageNum,
		Total:   total,
	})
}

func convertToSpordleSurfaceResponse(s model.SpordleSurface) models.SpordleSurfaceResponse {
	return models.SpordleSurfaceResponse{
		ID:                  s.ID,
		VenueID:             s.VenueID,
		VenueName:           s.VenueName,
		VenueAddress:        s.VenueAddress,
		VenueCity:           s.VenueCity,
		VenueRegion:         s.VenueRegion,
		VenueCountry:        s.VenueCountry,
		VenueAlias:          s.VenueAlias,
		VenueLatitude: func() float64 {
			if s.VenueLatitude != nil {
				return *s.VenueLatitude
			}
			return 0
		}(),
		VenueLongitude: func() float64 {
			if s.VenueLongitude != nil {
				return *s.VenueLongitude
			}
			return 0
		}(),
		VenuePostalCode:     s.VenuePostalCode,
		SurfaceID:           s.SurfaceID,
		SurfaceName:         s.SurfaceName,
		SurfaceAlias:        s.SurfaceAlias,
		SurfaceSports:       s.SurfaceSports,
		SurfaceTimeZone:     s.SurfaceTimeZone,
		SurfaceType:         s.SurfaceType,
		SurfaceSize:         s.SurfaceSize,
		LivebarnSurfaceID:   s.LivebarnSurfaceID,
		NumberOfGamesComing: s.NumberOfGamesComing,
		CreatedAt:           s.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:           s.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
