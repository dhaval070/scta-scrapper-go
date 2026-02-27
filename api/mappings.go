package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"surface-api/dao/model"
	"surface-api/models"

	"github.com/gin-gonic/gin"
)

// setSurface updates surface mapping for a site location
func (app *App) setSurface(c *gin.Context) {
	var input = &models.SiteLoc{}

	if err := c.BindJSON(input); err != nil {
		sendError(c, err)
		return
	}

	if input.SurfaceID == -1 {
		if err := app.db.Model(input).Where("site=? and location=?", input.Site, input.Location).Select("SurfaceID").Updates(input).Error; err != nil {
			sendError(c, err)
			return
		}
	} else {
		var surface = &model.Surface{}
		if err := app.db.Find(surface, input.SurfaceID).Error; err != nil {
			sendError(c, err)
			return
		}

		input.LocationID = surface.LocationID

		if err := app.db.Model(input).Where("site=? and location=?", input.Site, input.Location).Select("LocationID", "SurfaceID").Updates(input).Error; err != nil {
			sendError(c, err)
			return
		}
	}
	var result = []models.SiteLoc{}

	if err := app.db.Joins("LinkedSurface").Joins("LiveBarnLocation").Find(&result, "site=?", input.Site).Error; err != nil {
		sendError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// setLocation updates location mapping for a site location
func (app *App) setLocation(c *gin.Context) {
	var input = struct {
		LocationId int32  `json:"location_id"`
		Site       string `json:"site"`
		Location   string `json:"location"`
	}{}

	if err := c.BindJSON(&input); err != nil {
		sendError(c, err)
		return
	}

	err := app.db.Exec(`update sites_locations set location_id=? where site=? and location=?`, input.LocationId, input.Site, input.Location).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error updating location"})
		return
	}

	c.Request.URL.RawQuery = fmt.Sprintf("site=%s", input.Site)
	app.getSiteLoc(c)
}

// unsetMapping removes location or surface mapping for a site location
func (app *App) unsetMapping(c *gin.Context) {
	var input = &models.UnsetMappingInput{}

	if err := c.BindJSON(input); err != nil {
		sendError(c, err)
		return
	}

	updateFields := make(map[string]any)

	if input.Type == "location" {
		var current models.SiteLoc
		if err := app.db.Model(&models.SiteLoc{}).
			Where("site = ? AND location = ?", input.Site, input.Location).
			First(&current).Error; err != nil {
			sendError(c, err)
			return
		}

		updateFields["location_id"] = 0
		if current.SurfaceID != -1 {
			updateFields["surface_id"] = 0
		}
	} else {
		updateFields["surface_id"] = 0
	}

	if err := app.db.Model(&models.SiteLoc{}).
		Where("site = ? AND location = ?", input.Site, input.Location).
		Updates(updateFields).Error; err != nil {
		sendError(c, err)
		return
	}

	c.Request.URL.RawQuery = fmt.Sprintf("site=%s", input.Site)
	app.getSiteLoc(c)
}

// setMapping updates surface mapping for a site-specific mapping table
func (app *App) setMapping(c *gin.Context) {
	var input = &models.Mapping{}

	if err := c.BindJSON(input); err != nil {
		sendError(c, err)
		return
	}

	var table = input.Site + "_mappings"
	var surface = &model.Surface{}
	if err := app.db.Find(surface, input.SurfaceID).Error; err != nil {
		sendError(c, err)
		return
	}

	if err := app.db.Exec(fmt.Sprintf(`update %s set surface_id=? where location=?`, table), input.SurfaceID, input.Location).Error; err != nil {
		sendError(c, err)
		return
	}
	c.AddParam("site", input.Site)
	app.getMappings(c)
}

// getSiteLoc returns site location mappings for a given site
// @Summary Get paginated site location mappings
// @Description Returns paginated site location mappings for a given site with optional pagination parameters
// @Tags Mappings
// @Accept json
// @Produce json
// @Param site query string false "Site name"
// @Param location query string false "Filter locations starting with this value"
// @Param page query string false "Page number (default: 1)"
// @Param perPage query string false "Items per page (default: 10, max: 100)"
// @Success 200 {object} models.SiteLocResult
// @Security CookieAuth
// @Router /site-locations [get]
func (app *App) getSiteLoc(c *gin.Context) {
	site := c.Query("site")
	location := c.Query("location")
	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("perPage", "10")
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

	var total int64
	var result = []models.SiteLoc{}

	baseQuery := app.db.Model(&models.SiteLoc{}).Joins("LinkedSurface").Joins("LiveBarnLocation")

	if site != "" {
		baseQuery = baseQuery.Where("site=?", site)
	}

	if location != "" {
		baseQuery = baseQuery.Where("location LIKE ?", location+"%")
	}

	if err := baseQuery.Count(&total).Error; err != nil {
		sendError(c, err)
		return
	}

	if err := baseQuery.Offset(offset).Limit(perPageNum).Find(&result).Error; err != nil {
		sendError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.SiteLocResult{
		Data:    result,
		Page:    pageNum,
		PerPage: perPageNum,
		Total:   total,
	})
}

// getMHRLoc returns paginated MHR location mappings
func (app *App) getMHRLoc(c *gin.Context) {
	var result = []models.MhrLocation{}

	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("perPage", "10")
	name := c.Query("name")
	province := c.Query("province")

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

	var total int64

	baseQuery := app.db.Model(&models.MhrLocation{})

	if name != "" {
		baseQuery = baseQuery.Where(`rink_name like ?`, "%"+name+"%")
	}

	if province != "" {
		baseQuery = baseQuery.Where(`province like ?`, "%"+province+"%")
	}

	if err := baseQuery.Count(&total).Error; err != nil {
		sendError(c, err)
		return
	}

	if err := baseQuery.Joins("LinkedSurface").Joins("LiveBarnLocation").Offset(offset).Limit(perPageNum).Find(&result).Error; err != nil {
		sendError(c, err)
		return
	}

	c.JSON(http.StatusOK, models.MHRLocResult{
		Data:    result,
		Page:    pageNum,
		PerPage: perPageNum,
		Total:   total,
	})
}

// setMHRLoc updates location mapping for a MHR location
func (app *App) setMHRLoc(c *gin.Context) {
	var input = struct {
		LocationId int32 `json:"location_id"`
		MhrId      int   `json:"mhr_id"`
	}{}

	if err := c.BindJSON(&input); err != nil {
		sendError(c, err)
		return
	}

	err := app.db.Exec(`update mhr_locations set livebarn_location_id=? where mhr_id=?`, input.LocationId, input.MhrId).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error updating location"})
		return
	}

	c.Status(http.StatusOK)
}

// setMHRSurface updates surface mapping for a MHR location
func (app *App) setMHRSurface(c *gin.Context) {
	var input = &models.MhrLocation{}

	if err := c.BindJSON(input); err != nil {
		sendError(c, err)
		return
	}

	if input.LivebarnSurfaceId == -1 {
		if err := app.db.Exec(`update mhr_locations set livebarn_surface_id=? where mhr_id=?`, input.LivebarnSurfaceId, input.MhrID).Error; err != nil {
			sendError(c, err)
			return
		}
	} else {
		var surface = &model.Surface{}
		if err := app.db.Find(surface, input.LivebarnSurfaceId).Error; err != nil {
			sendError(c, err)
			return
		}

		input.LivebarnLocationId = int(surface.LocationID)

		if err := app.db.Exec(`update mhr_locations set livebarn_location_id=?, livebarn_surface_id=? where mhr_id=?`, input.LivebarnLocationId, input.LivebarnSurfaceId, input.MhrID).Error; err != nil {
			sendError(c, err)
			return
		}
	}
	c.Status(http.StatusOK)
}

// unsetMHRMapping removes location or surface mapping for a mhr location
func (app *App) unsetMHRMapping(c *gin.Context) {
	var input = &models.UnsetMHRMappingInput{}

	if err := c.BindJSON(input); err != nil {
		sendError(c, err)
		return
	}

	updateFields := make(map[string]any)

	if input.Type == "location" {
		var current models.MhrLocation
		if err := app.db.Model(&models.MhrLocation{}).
			Where("mhr_id = ?", input.MhrId).
			First(&current).Error; err != nil {
			sendError(c, err)
			return
		}

		updateFields["livebarn_location_id"] = 0
		if current.LivebarnSurfaceId != -1 {
			updateFields["livebarn_surface_id"] = 0
		}
	} else {
		updateFields["livebarn_surface_id"] = 0
	}

	if err := app.db.Model(&models.MhrLocation{}).
		Where("mhr_id = ?", input.MhrId).
		Updates(updateFields).Error; err != nil {
		sendError(c, err)
		return
	}
	c.Status(http.StatusOK)
}

// getMappings returns surface mappings for a given site
func (app *App) getMappings(c *gin.Context) {
	site := c.Param("site")
	var table = site + "_mappings"
	var result = []models.Mapping{}

	err := app.db.Raw(fmt.Sprintf(`SELECT
		"%s" as site,
		%s.location,
		%s.surface_id,
		surfaces.name as surface_name
		FROM %s
		LEFT JOIN surfaces ON %s.surface_id = surfaces.id`,
		site, table, table, table, table)).Scan(&result).Error

	if err != nil {
		sendError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// rampMappings returns RAMP location mappings for a province
func (app *App) rampMappings(c *gin.Context) {
	var province = c.Param("province")
	var result []models.RampLocation

	err := app.db.Raw(`SELECT
		a.rarid,
		a.name location, a.address, a.city, a.province_name, a.country, a.match_type,
		a.surface_id,
		b.name surface_name
		FROM RAMP_Locations a
		LEFT JOIN locations c ON c.id = a.location_id
		LEFT JOIN surfaces b ON b.id = a.surface_id
		WHERE a.province_name=?`, province).Scan(&result).Error

	if err != nil {
		sendError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// rampProvinces returns distinct provinces from RAMP_Locations
func (app *App) rampProvinces(c *gin.Context) {
	var result []sql.NullString

	err := app.db.Raw("select distinct(province_name) provinces from RAMP_Locations order by provinces").Scan(&result).Error

	if err != nil {
		sendError(c, err)
		return
	}
	var data []string
	for _, v := range result {
		if v.Valid {
			data = append(data, v.String)
		}
	}
	c.JSON(http.StatusOK, data)
}

// SetRampMappings updates surface mapping for a RAMP location
func (app *App) SetRampMappings(c *gin.Context) {
	var input = &models.SetRampSurfaceID{}

	if err := c.BindJSON(input); err != nil {
		sendError(c, err)
		return
	}

	var rec = &models.RampLocation{}

	if err := app.db.First(rec, "rarid=?", input.RarID).Error; err != nil {
		sendError(c, err)
		return
	}

	if err := app.db.Exec(`UPDATE RAMP_Locations SET surface_id=? where rarid=?`, input.SurfaceID, input.RarID).Error; err != nil {
		sendError(c, err)
		return
	}

	c.AddParam("province", input.Province)
	app.rampMappings(c)
}
