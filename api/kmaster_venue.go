package main

import (
	"errors"
	"net/http"
	"strconv"
	"surface-api/dao/model"
	"surface-api/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// getKmasterVenues retrieves all kmaster venue list records
// @Summary List kmaster venues
// @Description Returns a paginated list of all venues from the kmaster venue list. Use export=json to download all records as a JSON file with LiveBarn location and surface details.
// @Tags KmasterVenue
// @Accept json
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param perPage query int false "Items per page (default 10, max 100)"
// @Param country query string false "Filter by country"
// @Param state query string false "Filter by province/state"
// @Param name query string false "Filter by venue name (partial match)"
// @Param livebarn query bool false "Filter by livebarn venue match status"
// @Param export query string false "Export as JSON file download when set to 'json' (returns all matching records with LiveBarn location and surface details, no pagination)"
// @Success 200 {object} models.KVenueResult "Paginated venue list"
// @Success 200 {object} models.KmasterVenueExportResult "Downloaded venue list as kmaster-venues-{timestamp}.json"
// @Failure 500 {object} object "Internal server error"
// @Security CookieAuth
// @Router /kmaster-venues [get]
func (app *App) getKmasterVenues(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("perPage", "10")
	export := c.Query("export")

	pageNum, _ := strconv.Atoi(page)
	perPageNum, _ := strconv.Atoi(perPage)
	if pageNum < 1 {
		pageNum = 1
	}
	if perPageNum < 1 || perPageNum > 100 {
		perPageNum = 10
	}
	offset := (pageNum - 1) * perPageNum

	country := c.Query("country")
	state := c.Query("state")
	livebarn := c.Query("livebarn")
	name := c.Query("name")

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
	if name != "" {
		query = query.Where("venue_name LIKE ?", "%"+name+"%")
	}

	var total int64
	query.Count(&total)

	var venues []model.KmasterVenueList
	if export != "" {
		if err := query.Session(&gorm.Session{}).Order("id DESC").Find(&venues).Error; err != nil {
			sendError(c, err)
			return
		}
	} else {
		if err := query.Session(&gorm.Session{}).Offset(offset).Limit(perPageNum).Order("id DESC").Find(&venues).Error; err != nil {
			sendError(c, err)
			return
		}
	}

	// Collect all IDs for batch cross-checking
	var livebarnIDs []int
	var mhrIDs []int
	for _, v := range venues {
		if v.LivebarnVenueID != 0 {
			livebarnIDs = append(livebarnIDs, v.LivebarnVenueID)
		}
		if v.MhrVenueID != 0 {
			mhrIDs = append(mhrIDs, v.MhrVenueID)
		}
	}

	// Batch check livebarn_venue_id against locations table
	livebarnMatch := map[int]bool{}
	if export == "" && len(livebarnIDs) > 0 {
		var foundLocations []int
		app.db.Model(&model.Location{}).Select("id").Where("id IN ?", livebarnIDs).Find(&foundLocations)
		for _, id := range foundLocations {
			livebarnMatch[id] = true
		}
	}

	// Batch check mhr_venue_id against mhr_locations table
	mhrMatch := map[int]bool{}
	if len(mhrIDs) > 0 {
		var foundMhr []int
		app.db.Model(&model.MhrLocation{}).Select("mhr_id").Where("mhr_id IN ?", mhrIDs).Find(&foundMhr)
		for _, id := range foundMhr {
			mhrMatch[id] = true
		}
	}

	if export != "" {
		result := app.buildKmasterVenueExport(venues, mhrMatch, total)
		filename := "kmaster-venues-" + time.Now().Format("2006-01-02-150405") + ".json"
		c.Header("Content-Disposition", "attachment; filename="+filename)
		c.JSON(http.StatusOK, result)
		return
	}

	var response = []models.KmasterVenueListResponse{}
	for _, v := range venues {
		response = append(response, convertToKmasterVenueResponse(v, livebarnMatch[v.LivebarnVenueID], mhrMatch[v.MhrVenueID]))
	}

	c.JSON(http.StatusOK, models.KVenueResult{
		Data:    response,
		Page:    pageNum,
		PerPage: perPageNum,
		Total:   total,
	})
}

// getKmasterVenueByID retrieves a single kmaster venue record by ID
// @Summary Get kmaster venue by ID
// @Description Returns a single venue record from the kmaster venue list
// @Tags KmasterVenue
// @Accept json
// @Produce json
// @Param id path int true "Venue ID"
// @Success 200 {object} models.KmasterVenueListResponse
// @Failure 400 {object} object "Invalid ID"
// @Failure 404 {object} object "Venue not found"
// @Failure 500 {object} object "Internal server error"
// @Security CookieAuth
// @Router /kmaster-venues/{id} [get]
func (app *App) getKmasterVenueByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var venue model.KmasterVenueList
	if err := app.db.First(&venue, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Venue not found"})
			return
		}
		sendError(c, err)
		return
	}

	livebarnMatch := false
	if venue.LivebarnVenueID != 0 {
		var loc model.Location
		if err := app.db.First(&loc, venue.LivebarnVenueID).Error; err == nil {
			livebarnMatch = true
		}
	}
	mhrMatch := false
	if venue.MhrVenueID != 0 {
		var mhr model.MhrLocation
		if err := app.db.First(&mhr, venue.MhrVenueID).Error; err == nil {
			mhrMatch = true
		}
	}

	c.JSON(http.StatusOK, convertToKmasterVenueResponse(venue, livebarnMatch, mhrMatch))
}

// createKmasterVenue creates a new kmaster venue record
// @Summary Create kmaster venue
// @Description Creates a new venue record in the kmaster venue list
// @Tags KmasterVenue
// @Accept json
// @Produce json
// @Param input body models.KmasterVenueListInput true "Venue data"
// @Success 201 {object} models.KmasterVenueListResponse
// @Failure 400 {object} object "Invalid input"
// @Failure 500 {object} object "Internal server error"
// @Security CookieAuth
// @Router /kmaster-venues [post]
func (app *App) createKmasterVenue(c *gin.Context) {
	var input models.KmasterVenueListInput
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	venue := convertToKmasterVenueModel(input)
	if err := app.db.Create(&venue).Error; err != nil {
		sendError(c, err)
		return
	}

	livebarnMatch := false
	if venue.LivebarnVenueID != 0 {
		var loc model.Location
		if err := app.db.First(&loc, venue.LivebarnVenueID).Error; err == nil {
			livebarnMatch = true
		}
	}
	mhrMatch := false
	if venue.MhrVenueID != 0 {
		var mhr model.MhrLocation
		if err := app.db.First(&mhr, venue.MhrVenueID).Error; err == nil {
			mhrMatch = true
		}
	}

	c.JSON(http.StatusCreated, convertToKmasterVenueResponse(venue, livebarnMatch, mhrMatch))
}

// updateKmasterVenue updates an existing kmaster venue record
// @Summary Update kmaster venue
// @Description Updates an existing venue record in the kmaster venue list
// @Tags KmasterVenue
// @Accept json
// @Produce json
// @Param id path int true "Venue ID"
// @Param input body models.KmasterVenueListInput true "Updated venue data"
// @Success 200 {object} models.KmasterVenueListResponse
// @Failure 400 {object} object "Invalid input"
// @Failure 404 {object} object "Venue not found"
// @Failure 500 {object} object "Internal server error"
// @Security CookieAuth
// @Router /kmaster-venues/{id} [put]
func (app *App) updateKmasterVenue(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var input models.KmasterVenueListInput
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var venue model.KmasterVenueList
	if err := app.db.First(&venue, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Venue not found"})
			return
		}
		sendError(c, err)
		return
	}

	updated := convertToKmasterVenueModel(input)
	updated.ID = venue.ID
	updated.CreatedAt = venue.CreatedAt

	if err := app.db.Save(&updated).Error; err != nil {
		sendError(c, err)
		return
	}

	livebarnMatch := false
	if updated.LivebarnVenueID != 0 {
		var loc model.Location
		if err := app.db.First(&loc, updated.LivebarnVenueID).Error; err == nil {
			livebarnMatch = true
		}
	}
	mhrMatch := false
	if updated.MhrVenueID != 0 {
		var mhr model.MhrLocation
		if err := app.db.First(&mhr, updated.MhrVenueID).Error; err == nil {
			mhrMatch = true
		}
	}

	c.JSON(http.StatusOK, convertToKmasterVenueResponse(updated, livebarnMatch, mhrMatch))
}

// deleteKmasterVenue deletes a kmaster venue record
// @Summary Delete kmaster venue
// @Description Deletes a venue record from the kmaster venue list
// @Tags KmasterVenue
// @Accept json
// @Produce json
// @Param id path int true "Venue ID"
// @Success 200 {object} map[string]interface{} "Success message"
// @Failure 400 {object} object "Invalid ID"
// @Failure 404 {object} object "Venue not found"
// @Failure 500 {object} object "Internal server error"
// @Security CookieAuth
// @Router /kmaster-venues/{id} [delete]
func (app *App) deleteKmasterVenue(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var venue model.KmasterVenueList
	if err := app.db.First(&venue, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Venue not found"})
			return
		}
		sendError(c, err)
		return
	}

	if err := app.db.Delete(&venue).Error; err != nil {
		sendError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Venue deleted successfully"})
}

func (app *App) buildKmasterVenueExport(venues []model.KmasterVenueList, mhrMatch map[int]bool, total int64) models.KmasterVenueExportResult {
	livebarnMatch := map[int]bool{}

	var livebarnIDs []int
	for _, v := range venues {
		if v.LivebarnVenueID != 0 {
			livebarnIDs = append(livebarnIDs, v.LivebarnVenueID)
		}
	}

	locations := map[int]LocationWithSurfaces{}
	if len(livebarnIDs) > 0 {
		var locs []LocationWithSurfaces
		app.db.Where("id IN ?", livebarnIDs).Preload("Surfaces").Find(&locs)
		for _, loc := range locs {
			id := int(loc.ID)
			locations[id] = loc
			livebarnMatch[id] = true
		}
	}

	var exportData []models.KmasterVenueExportItem
	for _, v := range venues {
		item := models.KmasterVenueExportItem{
			KmasterVenueListResponse: convertToKmasterVenueResponse(v, livebarnMatch[v.LivebarnVenueID], mhrMatch[v.MhrVenueID]),
		}
		if loc, ok := locations[v.LivebarnVenueID]; ok {
			item.LivebarnLocation = &models.LivebarnLocationDetail{
				ID:         loc.ID,
				Name:       loc.Name,
				Address1:   loc.Address1,
				City:       loc.City,
				PostalCode: loc.PostalCode,
				ProvinceID: loc.ProvinceID,
				Surfaces:   loc.Surfaces,
			}
		}
		exportData = append(exportData, item)
	}

	return models.KmasterVenueExportResult{
		Data:  exportData,
		Total: total,
	}
}

func convertToKmasterVenueResponse(v model.KmasterVenueList, livebarnMatch bool, mhrMatch bool) models.KmasterVenueListResponse {
	return models.KmasterVenueListResponse{
		ID:                     v.ID,
		Validate:               v.Validate,
		LivebarnVenueID:        v.LivebarnVenueID,
		MhrVenueID:             v.MhrVenueID,
		VenueName:              v.VenueName,
		Surfaces:               v.Surfaces,
		City:                   v.City,
		RinkAddress:            v.RinkAddress,
		PostalCode:             v.PostalCode,
		ProvinceState:          v.ProvinceState,
		Country:                v.Country,
		CompanyNameAlt1:        v.CompanyNameAlt1,
		CompanyNameAlt2:        v.CompanyNameAlt2,
		CompanyNameAlt3:        v.CompanyNameAlt3,
		ParentCompany:          v.ParentCompany,
		VenueType:              v.VenueType,
		AccountStatus:          v.AccountStatus,
		StreamingPlatform:      v.StreamingPlatform,
		PhoneNumber:            v.PhoneNumber,
		Website:                v.Website,
		CreatedAt:              v.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:              v.UpdatedAt.Format("2006-01-02 15:04:05"),
		LivebarnVenueIDMatched: livebarnMatch,
		MhrVenueIDMatched:      mhrMatch,
	}
}

func convertToKmasterVenueModel(input models.KmasterVenueListInput) model.KmasterVenueList {
	v := model.KmasterVenueList{
		VenueName:         input.VenueName,
		City:              input.City,
		RinkAddress:       input.RinkAddress,
		PostalCode:        input.PostalCode,
		ProvinceState:     input.ProvinceState,
		Country:           input.Country,
		CompanyNameAlt1:   input.CompanyNameAlt1,
		CompanyNameAlt2:   input.CompanyNameAlt2,
		CompanyNameAlt3:   input.CompanyNameAlt3,
		ParentCompany:     input.ParentCompany,
		VenueType:         input.VenueType,
		AccountStatus:     input.AccountStatus,
		StreamingPlatform: input.StreamingPlatform,
		PhoneNumber:       input.PhoneNumber,
		Website:           input.Website,
	}
	if input.Validate != nil {
		v.Validate = *input.Validate
	}
	if input.LivebarnVenueID != nil {
		v.LivebarnVenueID = *input.LivebarnVenueID
	}
	if input.MhrVenueID != nil {
		v.MhrVenueID = *input.MhrVenueID
	}
	if input.Surfaces != nil {
		v.Surfaces = *input.Surfaces
	}
	return v
}
