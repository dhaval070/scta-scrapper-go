package main

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"
	"surface-api/dao/model"
	"surface-api/models"

	"github.com/gin-gonic/gin"
)

type SetMhrLbNotesInput struct {
	MhrID   int    `json:"mhr_id" binding:"required"`
	LbNotes string `json:"lb_notes"`
}

type GetMhrLbNotesResponse struct {
	MhrID   int    `json:"mhr_id"`
	LbNotes string `json:"lb_notes"`
}

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
// @Summary Get paginated MHR location mappings
// @Description Returns paginated MHR location mappings with optional filters. Include export query parameter to download CSV.
// @Tags Mappings
// @Accept json
// @Produce json
// @Param page query string false "Page number (default: 1)"
// @Param perPage query string false "Items per page (default: 10, max: 100)"
// @Param name query string false "Filter by rink name (partial match)"
// @Param province query string false "Filter by province (partial match)"
// @Param livebarn_location_id query string false "Filter by LiveBarn location mapping status (0 for unmapped, 1 for mapped)"
// @Param sort query string false "Sort column (mhr_id, created_at)"
// @Param order query string false "Sort direction (asc, desc)"
// @Param export query string false "Export as CSV when present (any non-empty value)"
// @Success 200 {object} models.MHRLocResult
// @Security CookieAuth
// @Router /mhr-locations [get]
func (app *App) getMHRLoc(c *gin.Context) {
	var result = []models.MhrLocation{}

	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("perPage", "10")
	name := c.Query("name")
	province := c.Query("province")
	livebarnLocationIdStr := c.Query("livebarn_location_id")
	export := c.Query("export")
	sort := c.DefaultQuery("sort", "mhr_id")
	order := c.DefaultQuery("order", "asc")

	baseQuery := app.db.Model(&models.MhrLocation{})

	if name != "" {
		baseQuery = baseQuery.Where(`rink_name like ?`, "%"+name+"%")
	}

	if province != "" {
		baseQuery = baseQuery.Where(`province like ?`, "%"+province+"%")
	}

	if livebarnLocationIdStr != "" {
		var livebarnLocationId int
		if _, err := fmt.Sscanf(livebarnLocationIdStr, "%d", &livebarnLocationId); err == nil {
			if livebarnLocationId == 0 {
				baseQuery = baseQuery.Where("livebarn_location_id = ?", 0)
			} else if livebarnLocationId == 1 {
				baseQuery = baseQuery.Where("livebarn_location_id != ?", 0)
			}
		}
	}

	allowedSorts := map[string]bool{"mhr_id": true, "created_at": true}
	if !allowedSorts[sort] {
		sort = "mhr_id"
	}
	allowedOrders := map[string]bool{"asc": true, "desc": true}
	if !allowedOrders[order] {
		order = "asc"
	}
	baseQuery = baseQuery.Order(fmt.Sprintf("%s %s", sort, order))

	if export != "" {
		if err := baseQuery.Joins("LiveBarnLocation").Find(&result).Error; err != nil {
			sendError(c, err)
			return
		}

		var b = &bytes.Buffer{}
		w := csv.NewWriter(b)
		w.Write([]string{"MhrID", "RinkName", "Aka", "Address", "Phone", "Website", "Notes", "LivebarnInstalled", "Province", "LivebarnLocationId", "LivebarnLocationName", "HomeTeams", "Livebarn Notes"})

		for _, row := range result {
			aka := ""
			if row.Aka != nil {
				aka = *row.Aka
			}
			phone := ""
			if row.Phone != nil {
				phone = *row.Phone
			}
			website := ""
			if row.Website != nil {
				website = *row.Website
			}
			notes := ""
			if row.Notes != nil {
				notes = *row.Notes
			}
			livebarnLocationName := ""
			if row.LiveBarnLocation.Name != "" {
				livebarnLocationName = row.LiveBarnLocation.Name
			}
			homeTeamsStr := ""
			if row.HomeTeams != nil {
				var labels []string
				for _, ht := range row.HomeTeams {
					if label, ok := ht["label"]; ok {
						labels = append(labels, label)
					}
				}
				homeTeamsStr = strings.Join(labels, ", ")
			}
			w.Write([]string{
				fmt.Sprint(row.MhrID),
				row.RinkName,
				aka,
				row.Address,
				phone,
				website,
				notes,
				fmt.Sprint(row.LivebarnInstalled),
				row.Province,
				fmt.Sprint(row.LivebarnLocationId),
				livebarnLocationName,
				homeTeamsStr,
				row.LbNotes,
			})
		}
		w.Flush()

		c.Writer.Header().Add("content-type", "text/csv")
		c.Writer.Header().Add("content-disposition", "attachment;filename=mhr_locations.csv")
		c.Writer.Write(b.Bytes())
		return
	}

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

// setMhrLbNotes updates lb_notes for a MHR location
// @Summary Update LiveBarn notes for MHR location
// @Description Updates the lb_notes field for a specific MHR location
// @Tags Mappings
// @Accept json
// @Produce json
// @Param input body SetMhrLbNotesInput true "MHR ID and LiveBarn notes"
// @Success 200 {object} gin.H "Success message"
// @Security CookieAuth
// @Router /mhr-lb-notes [post]
func (app *App) setMhrLbNotes(c *gin.Context) {
	var input SetMhrLbNotesInput

	if err := c.BindJSON(&input); err != nil {
		sendError(c, err)
		return
	}

	if err := app.db.Exec(`UPDATE mhr_locations SET lb_notes = ? WHERE mhr_id = ?`, input.LbNotes, input.MhrID).Error; err != nil {
		sendError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "LiveBarn notes updated successfully",
		"mhr_id":  input.MhrID,
	})
}

// getMhrLbNotes retrieves lb_notes for a MHR location
// @Summary Get LiveBarn notes for MHR location
// @Description Retrieves the lb_notes field for a specific MHR location
// @Tags Mappings
// @Accept json
// @Produce json
// @Param mhr_id path int true "MHR location ID"
// @Success 200 {object} GetMhrLbNotesResponse "LiveBarn notes"
// @Failure 404 {object} gin.H "MHR location not found"
// @Security CookieAuth
// @Router /mhr-lb-notes/{mhr_id} [get]
func (app *App) getMhrLbNotes(c *gin.Context) {
	mhrIDStr := c.Param("mhr_id")
	var mhrID int
	if _, err := fmt.Sscanf(mhrIDStr, "%d", &mhrID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid MHR ID format",
		})
		return
	}

	var result GetMhrLbNotesResponse

	if err := app.db.Raw(`SELECT mhr_id, lb_notes FROM mhr_locations WHERE mhr_id = ?`, mhrID).Scan(&result).Error; err != nil {
		sendError(c, err)
		return
	}

	// Check if record was found (Scan doesn't return ErrRecordNotFound for Raw queries)
	if result.MhrID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "MHR location not found",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
