package main

import (
	"fmt"
	"net/http"
	"surface-api/dao/model"
	"surface-api/models"

	"github.com/gin-gonic/gin"
)

// LocationWithSurfaces represents a location with its associated surfaces
type LocationWithSurfaces struct {
	model.Location
	Surfaces []model.Surface `json:"surfaces" gorm:"foreignKey:LocationID;references:ID"`
}

// getSurfaces returns list of all surfaces with location details
func (app *App) getSurfaces(c *gin.Context) {
	var surfaces = []models.SurfaceResult{}
	province := c.Query("province")

	var err error

	err = app.db.Raw(`SELECT
			a.id,
			a.location_id,
			a.name,
			a.sports,
			l.name location_name,
			l.city location_city,
			l.address1 location_address
		FROM
			surfaces a
		INNER JOIN locations l ON a.location_id = l.id
		INNER JOIN provinces p ON l.province_id = p.id
		WHERE p.province_name = ? ORDER BY l.name`, province).Scan(&surfaces).Error

	if err != nil {
		sendError(c, err)
	}
	c.JSON(http.StatusOK, surfaces)
}

// getProvinces returns list of all provinces in Canada
func (app *App) getProvinces(c *gin.Context) {
	var result = []struct {
		Id           int32  `json:"id"`
		ProvinceName string `json:"province_name"`
	}{}

	err := app.db.Raw(`select id, province_name from provinces order by province_name`).
		Scan(&result).Error

	if err != nil {
		sendError(c, err)
	}

	c.JSON(http.StatusOK, result)
}

// getSites returns list of all configured sites
func (app *App) getSites(c *gin.Context) {
	var sites = []model.SitesConfig{}

	if err := app.db.Order("display_name asc").Find(&sites).Error; err != nil {
		sendError(c, err)
		return
	}
	c.JSON(http.StatusOK, sites)
}

// getLocations returns paginated list of locations with their surfaces
func (app *App) getLocations(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("perPage", "10")
	name := c.Query("name")
	postalCode := c.Query("postal_code")

	var pageNum, perPageNum int
	fmt.Sscanf(page, "%d", &pageNum)
	fmt.Sscanf(perPage, "%d", &perPageNum)

	if pageNum < 1 {
		pageNum = 1
	}
	if perPageNum < 1 || perPageNum > 100 {
		perPageNum = 10
	}

	offset := (pageNum - 1) * perPageNum

	baseQuery := app.db.Model(&model.Location{}).Order("locations.name")

	if name != "" {
		baseQuery = baseQuery.Where("name LIKE ?", "%"+name+"%")
	}
	if postalCode != "" {
		baseQuery = baseQuery.Where("postal_code LIKE ?", "%"+postalCode+"%")
	}

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		sendError(c, err)
		return
	}

	var result []LocationWithSurfaces
	if err := baseQuery.Preload("Surfaces").Limit(perPageNum).Offset(offset).Find(&result).Error; err != nil {
		sendError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    result,
		"page":    pageNum,
		"perPage": perPageNum,
		"total":   total,
	})
}
