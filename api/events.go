package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"surface-api/dao/model"
	"surface-api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UpdateEventInput struct {
	SurfaceID    int32 `json:"surface_id"`
	UpdateFuture bool  `json:"update_future"`
}

// getEvents returns paginated list of events with optional filters
// @Summary Get paginated events
// @Description Returns paginated list of events with optional filters for site, surface, date range, and claim status
// @Tags Events
// @Accept json
// @Produce json
// @Param page query string false "Page number (default: 1)"
// @Param perPage query string false "Items per page (default: 10, max: 100)"
// @Param site query string false "Filter by site name"
// @Param surface_id query string false "Filter by surface ID"
// @Param start_date query string false "Filter by start date (YYYY-MM-DD)"
// @Param end_date query string false "Filter by end date (YYYY-MM-DD)"
// @Param export query string false "Export as CSV when present (any non-empty value)"
// @Param claim_status query string false "Filter by claim status (success or error)"
// @Success 200 {object} models.EventsResult
// @Failure 400 {object} map[string]interface{} "Invalid claim_status value"
// @Security CookieAuth
// @Router /events [get]
func (app *App) getEvents(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("perPage", "10")
	site := c.Query("site")
	surfaceID := c.Query("surface_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	export := c.Query("export")
	claimStatus := c.Query("claim_status")

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

	baseQuery := app.db.Table("events").
		Select("events.*, claim_api_log.status AS claim_status, claim_api_log.http_status_code AS claim_http_status_code, claim_api_log.error_message AS claim_error_message, claim_api_log.created_at AS claim_created_at, claim_api_log.updated_at AS claim_updated_at").
		Joins("LEFT JOIN claim_api_log ON claim_api_log.event_id = events.event_id").
		Order("events.datetime DESC")

	if site != "" {
		baseQuery = baseQuery.Where("events.site = ?", site)
	}
	if surfaceID != "" {
		baseQuery = baseQuery.Where("events.surface_id = ?", surfaceID)
	}
	if startDate != "" {
		baseQuery = baseQuery.Where("events.datetime >= ?", startDate)
	}
	if endDate != "" {
		baseQuery = baseQuery.Where("events.datetime <= ?", endDate)
	}
	if claimStatus != "" {
		if claimStatus == "success" {
			baseQuery = baseQuery.Where("claim_api_log.status = 1")
		} else if claimStatus == "error" {
			baseQuery = baseQuery.Where("claim_api_log.status = 0")
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid claim_status value (must be 'success' or 'error')"})
			return
		}
	}

	var total int64
	if export == "" {
		if err := baseQuery.Count(&total).Error; err != nil {
			sendError(c, err)
			return
		}
	}

	var result []models.EventWithClaim
	if export != "" {
		if err := baseQuery.Scan(&result).Error; err != nil {
			sendError(c, err)
			return
		}

		var b = &bytes.Buffer{}
		w := csv.NewWriter(b)
		w.Write([]string{"ID", "Site", "Date/Time", "Home Team", "Guest Team", "Location", "Division", "Event ID", "Surface ID", "Claim Status", "HTTP Status Code", "Error Message", "Claim Created At", "Claim Updated At"})

		for _, row := range result {
			claimStatusStr := ""
			if row.ClaimStatus != nil {
				claimStatusStr = fmt.Sprint(*row.ClaimStatus)
			}
			claimHTTPCode := ""
			if row.ClaimHTTPStatusCode != nil {
				claimHTTPCode = fmt.Sprint(*row.ClaimHTTPStatusCode)
			}
			claimErrorMsg := ""
			if row.ClaimErrorMessage != nil {
				claimErrorMsg = *row.ClaimErrorMessage
			}
			claimCreatedAt := ""
			if row.ClaimCreatedAt != nil {
				claimCreatedAt = row.ClaimCreatedAt.Format("2006-01-02 15:04")
			}
			claimUpdatedAt := ""
			if row.ClaimUpdatedAt != nil {
				claimUpdatedAt = row.ClaimUpdatedAt.Format("2006-01-02 15:04")
			}

			w.Write([]string{
				fmt.Sprint(row.ID),
				row.Site,
				row.Datetime.Format("2006-01-02 15:04"),
				row.HomeTeam,
				row.GuestTeam,
				row.Location,
				row.Division,
				row.EventID,
				fmt.Sprint(row.SurfaceID),
				claimStatusStr,
				claimHTTPCode,
				claimErrorMsg,
				claimCreatedAt,
				claimUpdatedAt,
			})
		}
		w.Flush()

		c.Writer.Header().Add("content-type", "text/csv")
		c.Writer.Header().Add("content-disposition", "attachment;filename=events_report.csv")
		c.Writer.Write(b.Bytes())

		return
	}
	if err := baseQuery.Limit(perPageNum).Offset(offset).Scan(&result).Error; err != nil {
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

// updateEvent updates the surface_id and location_id of an event
// @Summary Update event surface assignment
// @Description Updates the surface_id (and location_id) for a single event or all future events at the same location
// @Tags Events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Param input body UpdateEventInput true "Event update parameters"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{} "Event or Surface not found"
// @Security CookieAuth
// @Router /events/{id} [put]
func (app *App) updateEvent(c *gin.Context) {
	id := c.Param("id")
	var input UpdateEventInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var event model.Event
	if err := app.db.First(&event, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}
		sendError(c, err)
		return
	}

	var locationID int32 = 0

	if input.SurfaceID != 0 {
		var surface model.Surface
		if err := app.db.First(&surface, input.SurfaceID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Surface not found"})
				return
			}
			sendError(c, err)
			return
		}
		locationID = surface.LocationID
	}

	if input.UpdateFuture {
		result := app.db.Model(&model.Event{}).
			Where("site = ? AND location = ? AND datetime >= ?", event.Site, event.Location, event.Datetime).
			Updates(map[string]interface{}{
				"surface_id":  input.SurfaceID,
				"location_id": locationID,
			})

		if result.Error != nil {
			sendError(c, result.Error)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       "Events updated successfully",
			"updated_count": result.RowsAffected,
			"surface_id":    input.SurfaceID,
			"location_id":   locationID,
			"update_future": true,
		})
	} else {
		event.SurfaceID = input.SurfaceID
		event.LocationID = locationID

		if err := app.db.Save(&event).Error; err != nil {
			sendError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":     "Event updated successfully",
			"event":       event,
			"surface_id":  input.SurfaceID,
			"location_id": locationID,
		})
	}
}
