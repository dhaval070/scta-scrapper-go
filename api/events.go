package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"surface-api/dao/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UpdateEventInput struct {
	SurfaceID    int32 `json:"surface_id"`
	UpdateFuture bool  `json:"update_future"`
}

// getEvents returns paginated list of events with optional filters
func (app *App) getEvents(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("perPage", "10")
	site := c.Query("site")
	surfaceID := c.Query("surface_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	export := c.Query("export")

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

	baseQuery := app.db.Model(&model.Event{}).Order("datetime DESC")

	if site != "" {
		baseQuery = baseQuery.Where("site = ?", site)
	}
	if surfaceID != "" {
		baseQuery = baseQuery.Where("surface_id = ?", surfaceID)
	}
	if startDate != "" {
		baseQuery = baseQuery.Where("datetime >= ?", startDate)
	}
	if endDate != "" {
		baseQuery = baseQuery.Where("datetime <= ?", endDate)
	}

	var total int64
	if export == "" {
		if err := baseQuery.Count(&total).Error; err != nil {
			sendError(c, err)
			return
		}
	}

	var result []model.Event
	if export != "" {
		if err := baseQuery.Find(&result).Error; err != nil {
			sendError(c, err)
			return
		}

		var b = &bytes.Buffer{}
		w := csv.NewWriter(b)
		w.Write([]string{"ID", "Site", "Date/Time", "Home Team", "Gues Team", "Location", "Division", "Surface ID"})

		for _, row := range result {
			w.Write([]string{
				fmt.Sprint(row.ID),
				row.Site,
				row.Datetime.Format("2006-01-02 15:04"),
				row.HomeTeam,
				row.GuestTeam,
				row.Location,
				row.Division,
				fmt.Sprint(row.SurfaceID),
			})
		}
		w.Flush()

		c.Writer.Header().Add("content-type", "text/csv")
		c.Writer.Header().Add("content-disposition", "attachment;filename=events_report.csv")
		c.Writer.Write(b.Bytes())

		return
	}
	if err := baseQuery.Limit(perPageNum).Offset(offset).Find(&result).Error; err != nil {
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
