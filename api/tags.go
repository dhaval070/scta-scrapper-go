package main

import (
	"errors"
	"net/http"
	"strconv"
	"surface-api/dao/model"
	"surface-api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// listTags returns all tags with optional name search
// @Summary List all tags
// @Description Returns all tags, optionally filtered by name search
// @Tags Tags
// @Accept json
// @Produce json
// @Param q query string false "Search tags by name (partial match)"
// @Success 200 {array} model.Tag
// @Security CookieAuth
// @Router /tags [get]
func (app *App) listTags(c *gin.Context) {
	q := c.Query("q")

	query := app.db.Model(&model.Tag{})
	if q != "" {
		query = query.Where("name LIKE ?", "%"+q+"%")
	}

	var tags []model.Tag
	if err := query.Order("name ASC").Find(&tags).Error; err != nil {
		sendError(c, err)
		return
	}

	c.JSON(http.StatusOK, tags)
}

// createTag creates a new tag
// @Summary Create a tag
// @Description Creates a new tag with the given name, optional color and description
// @Tags Tags
// @Accept json
// @Produce json
// @Param input body models.CreateTagInput true "Tag details"
// @Success 201 {object} model.Tag
// @Security CookieAuth
// @Router /tags [post]
func (app *App) createTag(c *gin.Context) {
	var input models.CreateTagInput
	if err := c.BindJSON(&input); err != nil {
		sendError(c, err)
		return
	}

	tag := model.Tag{
		Name:        input.Name,
		Color:       "",
		Description: "",
	}
	if input.Color != nil {
		tag.Color = *input.Color
	}
	if input.Description != nil {
		tag.Description = *input.Description
	}

	if err := app.db.Create(&tag).Error; err != nil {
		sendError(c, err)
		return
	}

	c.JSON(http.StatusCreated, tag)
}

// updateTag updates an existing tag
// @Summary Update a tag
// @Description Updates the name, color, or description of an existing tag
// @Tags Tags
// @Accept json
// @Produce json
// @Param id path int true "Tag ID"
// @Param input body models.UpdateTagInput true "Tag fields to update"
// @Success 200 {object} model.Tag
// @Security CookieAuth
// @Router /tags/{id} [put]
func (app *App) updateTag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}

	var tag model.Tag
	if err := app.db.First(&tag, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
			return
		}
		sendError(c, err)
		return
	}

	var input models.UpdateTagInput
	if err := c.BindJSON(&input); err != nil {
		sendError(c, err)
		return
	}

	updates := make(map[string]any)
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Color != nil {
		updates["color"] = *input.Color
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	if err := app.db.Model(&tag).Updates(updates).Error; err != nil {
		sendError(c, err)
		return
	}

	// Re-fetch to get updated values
	app.db.First(&tag, id)
	c.JSON(http.StatusOK, tag)
}

// deleteTag deletes a tag
// @Summary Delete a tag
// @Description Deletes a tag and removes all its associations (via CASCADE)
// @Tags Tags
// @Accept json
// @Produce json
// @Param id path int true "Tag ID"
// @Success 204 "No Content"
// @Security CookieAuth
// @Router /tags/{id} [delete]
func (app *App) deleteTag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}

	result := app.db.Delete(&model.Tag{}, id)
	if result.Error != nil {
		sendError(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

// getSiteLocationTags returns tags for a specific site-location
// @Summary Get tags for a site location
// @Description Returns all tags assigned to a specific site-location pair
// @Tags Tags
// @Accept json
// @Produce json
// @Param site query string true "Site name"
// @Param location query string true "Location name"
// @Success 200 {array} model.Tag
// @Failure 400 {object} map[string]interface{} "Missing required query parameters"
// @Security CookieAuth
// @Router /site-location-tags [get]
func (app *App) getSiteLocationTags(c *gin.Context) {
	site := c.Query("site")
	location := c.Query("location")

	if site == "" || location == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "site and location query parameters are required"})
		return
	}

	var tags []model.Tag
	if err := app.db.Model(&model.Tag{}).
		Joins("JOIN sites_location_tags ON sites_location_tags.tag_id = tags.id").
		Where("sites_location_tags.site = ? AND sites_location_tags.location = ?", site, location).
		Order("tags.name ASC").
		Find(&tags).Error; err != nil {
		sendError(c, err)
		return
	}

	c.JSON(http.StatusOK, tags)
}

// addSiteLocationTags adds tags to a site-location
// @Summary Add tags to a site location
// @Description Associates tags (by ID) with a specific site-location pair
// @Tags Tags
// @Accept json
// @Produce json
// @Param site query string true "Site name"
// @Param location query string true "Location name"
// @Param input body models.AddTagsToSiteLocationInput true "Tag IDs to add"
// @Success 201 {array} model.Tag
// @Failure 400 {object} map[string]interface{} "Missing or invalid parameters"
// @Security CookieAuth
// @Router /site-location-tags [post]
func (app *App) addSiteLocationTags(c *gin.Context) {
	site := c.Query("site")
	location := c.Query("location")

	if site == "" || location == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "site and location query parameters are required"})
		return
	}

	var input models.AddTagsToSiteLocationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(input.TagIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tag_ids must not be empty"})
		return
	}

	// Verify site-location exists
	var count int64
	if err := app.db.Model(&model.SitesLocation{}).
		Where("site = ? AND location = ?", site, location).
		Count(&count).Error; err != nil {
		sendError(c, err)
		return
	}
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "site location not found"})
		return
	}

	// Verify all tags exist
	var tagCount int64
	if err := app.db.Model(&model.Tag{}).Where("id IN ?", input.TagIDs).Count(&tagCount).Error; err != nil {
		sendError(c, err)
		return
	}
	if tagCount != int64(len(input.TagIDs)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "one or more tag_ids not found"})
		return
	}

	// Insert join records, ignoring duplicates
	for _, tagID := range input.TagIDs {
		app.db.Exec(
			"INSERT IGNORE INTO sites_location_tags (site, location, tag_id) VALUES (?, ?, ?)",
			site, location, tagID,
		)
	}

	// Return updated tags
	var tags []model.Tag
	app.db.Model(&model.Tag{}).
		Joins("JOIN sites_location_tags ON sites_location_tags.tag_id = tags.id").
		Where("sites_location_tags.site = ? AND sites_location_tags.location = ?", site, location).
		Order("tags.name ASC").
		Find(&tags)

	c.JSON(http.StatusCreated, tags)
}

// removeSiteLocationTag removes a tag from a site-location
// @Summary Remove a tag from a site location
// @Description Removes an association between a tag and a site-location
// @Tags Tags
// @Accept json
// @Produce json
// @Param site query string true "Site name"
// @Param location query string true "Location name"
// @Param tag_id query int true "Tag ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]interface{} "Missing or invalid parameters"
// @Security CookieAuth
// @Router /site-location-tags [delete]
func (app *App) removeSiteLocationTag(c *gin.Context) {
	site := c.Query("site")
	location := c.Query("location")
	tagIDStr := c.Query("tag_id")

	if site == "" || location == "" || tagIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "site, location, and tag_id query parameters are required"})
		return
	}

	tagID, err := strconv.Atoi(tagIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}

	result := app.db.Exec(
		"DELETE FROM sites_location_tags WHERE site = ? AND location = ? AND tag_id = ?",
		site, location, tagID,
	)
	if result.Error != nil {
		sendError(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag association not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

// getSiteTags returns tags for a specific site
// @Summary Get tags for a site
// @Description Returns all tags assigned to a specific site
// @Tags Tags
// @Accept json
// @Produce json
// @Param site_name query string true "Site name"
// @Success 200 {array} model.Tag
// @Failure 400 {object} map[string]interface{} "Missing required query parameters"
// @Security CookieAuth
// @Router /sites-tags [get]
func (app *App) getSiteTags(c *gin.Context) {
	siteName := c.Query("site_name")

	if siteName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "site_name query parameter is required"})
		return
	}

	var tags []model.Tag
	if err := app.db.Model(&model.Tag{}).
		Joins("JOIN sites_tags ON sites_tags.tag_id = tags.id").
		Where("sites_tags.site_name = ?", siteName).
		Order("tags.name ASC").
		Find(&tags).Error; err != nil {
		sendError(c, err)
		return
	}

	c.JSON(http.StatusOK, tags)
}

// addSiteTags adds tags to a site
// @Summary Add tags to a site
// @Description Associates tags (by ID) with a specific site
// @Tags Tags
// @Accept json
// @Produce json
// @Param site_name query string true "Site name"
// @Param input body models.AddTagsToSiteInput true "Tag IDs to add"
// @Success 201 {array} model.Tag
// @Failure 400 {object} map[string]interface{} "Missing or invalid parameters"
// @Security CookieAuth
// @Router /sites-tags [post]
func (app *App) addSiteTags(c *gin.Context) {
	siteName := c.Query("site_name")

	if siteName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "site_name query parameter is required"})
		return
	}

	var input models.AddTagsToSiteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(input.TagIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tag_ids must not be empty"})
		return
	}

	// Verify site exists
	var count int64
	if err := app.db.Model(&model.SitesConfig{}).
		Where("site_name = ?", siteName).
		Count(&count).Error; err != nil {
		sendError(c, err)
		return
	}
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "site not found"})
		return
	}

	// Verify all tags exist
	var tagCount int64
	if err := app.db.Model(&model.Tag{}).Where("id IN ?", input.TagIDs).Count(&tagCount).Error; err != nil {
		sendError(c, err)
		return
	}
	if tagCount != int64(len(input.TagIDs)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "one or more tag_ids not found"})
		return
	}

	// Insert join records, ignoring duplicates
	for _, tagID := range input.TagIDs {
		app.db.Exec(
			"INSERT IGNORE INTO sites_tags (site_name, tag_id) VALUES (?, ?)",
			siteName, tagID,
		)
	}

	// Return updated tags
	var tags []model.Tag
	app.db.Model(&model.Tag{}).
		Joins("JOIN sites_tags ON sites_tags.tag_id = tags.id").
		Where("sites_tags.site_name = ?", siteName).
		Order("tags.name ASC").
		Find(&tags)

	c.JSON(http.StatusCreated, tags)
}

// removeSiteTag removes a tag from a site
// @Summary Remove a tag from a site
// @Description Removes an association between a tag and a site
// @Tags Tags
// @Accept json
// @Produce json
// @Param site_name query string true "Site name"
// @Param tag_id query int true "Tag ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]interface{} "Missing or invalid parameters"
// @Security CookieAuth
// @Router /sites-tags [delete]
func (app *App) removeSiteTag(c *gin.Context) {
	siteName := c.Query("site_name")
	tagIDStr := c.Query("tag_id")

	if siteName == "" || tagIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "site_name and tag_id query parameters are required"})
		return
	}

	tagID, err := strconv.Atoi(tagIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}

	result := app.db.Exec(
		"DELETE FROM sites_tags WHERE site_name = ? AND tag_id = ?",
		siteName, tagID,
	)
	if result.Error != nil {
		sendError(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag association not found"})
		return
	}

	c.Status(http.StatusNoContent)
}
