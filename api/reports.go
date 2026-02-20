package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"surface-api/dao/model"
	"surface-api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

type RinkReportItem struct {
	EDate      string         `json:"edate"`
	Rink       string         `json:"rink"`
	LocationID int32          `json:"location_id"`
	City       string         `json:"city"`
	Province   string         `json:"province"`
	JsonReport map[string]int `json:"json_report"`
	Total      int32          `json:"total"`
}

// RinkReportResponse describes the full API response for /rink-report
type RinkReportResponse struct {
	Data      []RinkReportItem `json:"data"`
	Page      int              `json:"page"`
	PerPage   int              `json:"perPage"`
	Total     int64            `json:"total"`
	StartDate string           `json:"start_date"`
	EndDate   string           `json:"end_date"`
}

// surfaceReport returns paginated surface usage report
func (app *App) surfaceReport(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("perPage", "10")
	locationName := c.Query("location_name")

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

	whereClause := ""
	var args []any
	if locationName != "" {
		whereClause = " WHERE l.name LIKE ?"
		args = append(args, "%"+locationName+"%")
	}

	countQuery := `SELECT COUNT(*) FROM (
		SELECT e.surface_id,
			any_value(l.name) location_name,
			any_value(s.name) surface_name,
			date(e.datetime)
		FROM events e 
		JOIN surfaces s ON e.surface_id=s.id 
		JOIN locations l ON l.id=s.location_id` +
		whereClause +
		`
		GROUP BY e.surface_id, date(e.datetime)
	) AS subquery`

	var total int64
	if err := app.db.Raw(countQuery, args...).Scan(&total).Error; err != nil {
		sendError(c, err)
		return
	}

	query := `SELECT
		e.surface_id,
		any_value(s.location_id) as location_id,
		any_value(l.name) location_name,
		any_value(s.name) surface_name,
		any_value(date_format(e.datetime, "%W")) day_of_week,
		date_format(min(e.datetime), "%Y-%m-%d %T") start_time,
		date_format(max( date_add(e.datetime, INTERVAL 150 minute)), "%Y-%m-%d %T") end_time
	FROM
		events e JOIN surfaces s on e.surface_id=s.id JOIN locations l on l.id=s.location_id` +
		whereClause +
		`
	GROUP BY e.surface_id, date(e.datetime)
	ORDER BY location_name, surface_name, day_of_week,  start_time, end_time
	LIMIT ? OFFSET ?`

	queryArgs := append(args, perPageNum, offset)
	var result []models.SurfaceReport
	if err := app.db.Raw(query, queryArgs...).Scan(&result).Error; err != nil {
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

// downloadReportCSV returns surface report as CSV
func (app *App) downloadReportCSV(c *gin.Context) {
	locationName := c.Query("location_name")

	whereClause := ""
	var args []interface{}
	if locationName != "" {
		whereClause = " WHERE l.name LIKE ?"
		args = append(args, "%"+locationName+"%")
	}

	query := `SELECT
		e.surface_id,
		any_value(s.location_id),
		any_value(l.name) location_name,
		any_value(s.name) surface_name,
		any_value(date_format(e.datetime, "%W")) day_of_week,
		date_format(min(e.datetime), "%Y-%m-%d %T") start_time,
		date_format(max( date_add(e.datetime, INTERVAL 150 minute)), "%Y-%m-%d %T") end_time
	FROM
		events e JOIN surfaces s on e.surface_id=s.id JOIN locations l on l.id=s.location_id` +
		whereClause +
		`
	GROUP BY e.surface_id, date(e.datetime)
	ORDER BY location_name, surface_name, surface_id, day_of_week, start_time, end_time`

	var result []models.SurfaceReport
	if err := app.db.Raw(query, args...).Scan(&result).Error; err != nil {
		sendError(c, err)
		return
	}

	var b = &bytes.Buffer{}
	w := csv.NewWriter(b)

	w.Write([]string{
		"Surface ID", "Location Name", "Surface Name", "Day of Week", "Start Time", "End Time",
	})

	for _, row := range result {
		w.Write([]string{
			row.SurfaceID,
			row.LocationName,
			row.SurfaceName,
			row.DayOfWeek,
			row.StartTime,
			row.EndTime,
		})
	}
	w.Flush()

	c.Writer.Header().Add("content-type", "text/csv")
	c.Writer.Header().Add("content-disposition", "attachment;filename=surface_report.csv")
	c.Writer.Write(b.Bytes())
}

// rinkReport returns paginated rink usage report
func (app *App) rinkReport(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("perPage", "10")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	export := c.Query("export")
	site := c.Query("site")

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

	where := ""
	var args []any
	rink := c.Query("rink")
	province := c.Query("province")
	city := c.Query("city")
	if startDate != "" {
		where = " WHERE edate >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		if where == "" {
			where = " WHERE edate <= ?"
			args = append(args, endDate)
		} else {
			where = where + " AND edate <= ?"
			args = append(args, endDate)
		}
	}
	if rink != "" {
		if where == "" {
			where = " WHERE l.name LIKE ?"
		} else {
			where = where + " AND l.name LIKE ?"
		}
		args = append(args, "%"+rink+"%")
	}
	if province != "" {
		if where == "" {
			where = " WHERE p.id = ?"
		} else {
			where = where + " AND p.id = ?"
		}
		args = append(args, province)
	}
	if city != "" {
		if where == "" {
			where = " WHERE l.city LIKE ?"
		} else {
			where = where + " AND l.city LIKE ?"
		}
		args = append(args, city+"%")
	}
	if site != "" {
		if where == "" {
			where = " WHERE e.site=?"
		} else {
			where = where + " AND e.site=?"
		}
		args = append(args, site)
	}

	countQuery := `with tbl as (
		select edate, l.name, location_id, l.city,p.province_name, site, count(*) cnt from events e
		inner join locations l on l.id=e.location_id
		inner join provinces p on p.id=l.province_id
		` + where + `
		group by edate, location_id, site
		) select count(edate) from tbl;`

	var total int64
	if err := app.db.Raw(countQuery, args...).Scan(&total).Error; err != nil {
		sendError(c, err)
		return
	}

	// Build query differently when exporting (no pagination)
	var query string
	var queryArgs []any
	if export != "" {
		query = `with tbl as (
			select edate, l.name, location_id, l.city,p.province_name, site, count(*) cnt from events e
			inner join locations l on l.id=e.location_id
			inner join provinces p on p.id=l.province_id
			` + where + `
			group by edate, location_id, site
			)
			select edate e_date,
			any_value(name) rink,
			tbl.location_id,
			any_value(city) city,
			any_value(province_name) province,
			json_objectagg(tbl.site, cnt) json_report,
			sum(cnt) total
			from
			tbl
			group by edate, location_id order by edate,name`
		queryArgs = args
	} else {
		query = `with tbl as (
			select edate, l.name, location_id, l.city,p.province_name, site, count(*) cnt from events e
			join locations l on l.id=e.location_id
			join provinces p on p.id=l.province_id
			` + where + `
			group by edate, location_id, site
			LIMIT ? OFFSET ?
			)
			select edate e_date,
			any_value(name) rink,
			tbl.location_id,
			any_value(city) city,
			any_value(province_name) province,
			json_objectagg(tbl.site, cnt) json_report,
			sum(cnt) total
			from
			tbl
			group by edate, location_id order by edate,name`
		queryArgs = append(args, perPageNum, offset)
	}

	var result = []struct {
		EDate      string         `json:"edate"`
		Rink       string         `json:"rink"`
		LocationID int32          `json:"location_id"`
		City       string         `json:"city"`
		Province   string         `json:"province"`
		JsonReport datatypes.JSON `json:"json_report"`
		Total      int32          `json:"total"`
	}{}

	if err := app.db.Raw(query, queryArgs...).Scan(&result).Error; err != nil {
		sendError(c, err)
		return
	}

	// If export param present, stream CSV without pagination
	if export != "" {
		var b = &bytes.Buffer{}
		w := csv.NewWriter(b)
		w.Write([]string{"Date", "Rink", "Location ID", "City", "Province", "Json Report", "Total"})
		for _, row := range result {
			w.Write([]string{
				row.EDate,
				row.Rink,
				fmt.Sprint(row.LocationID),
				row.City,
				row.Province,
				string(row.JsonReport),
				fmt.Sprint(row.Total),
			})
		}
		w.Flush()

		c.Writer.Header().Add("content-type", "text/csv")
		c.Writer.Header().Add("content-disposition", "attachment;filename=rink_report.csv")
		c.Writer.Write(b.Bytes())
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result, "page": pageNum, "perPage": perPageNum, "total": total, "start_date": startDate, "end_date": endDate})
}

// getEventsByDateRange returns events for a specific location within a date range
func (app *App) getEventsByDateRange(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	locationID := c.Query("location_id")

	if startDate == "" || endDate == "" || locationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "start_date, end_date, and location_id are required parameters",
		})
		return
	}

	var location model.Location
	if err := app.db.First(&location, locationID).Error; err != nil {
		sendError(c, err)
		return
	}

	var results []models.EventWithLocation
	if err := app.db.Table("events").
		Select("events.*, locations.name as location_name, surfaces.name as surface_name, sites_config.display_name as display_name").
		Joins("LEFT JOIN locations ON events.location_id = locations.id").
		Joins("LEFT JOIN surfaces ON events.surface_id = surfaces.id").
		Joins("LEFT JOIN sites_config ON events.site = sites_config.site_name").
		Where("events.datetime >= ? AND events.datetime <= ? AND events.location_id = ?", startDate, endDate, locationID).
		Order("events.datetime ASC").
		Scan(&results).Error; err != nil {
		sendError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":          results,
		"start_date":    startDate,
		"end_date":      endDate,
		"location_id":   locationID,
		"location_name": location.Name,
		"count":         len(results),
	})
}
