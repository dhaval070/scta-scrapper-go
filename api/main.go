package main

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"surface-api/dao/model"
	"surface-api/models"
	"sync"
	"time"

	_ "surface-api/docs"

	"github.com/astaxie/beego/session"
	_ "github.com/astaxie/beego/session/mysql"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/datatypes"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// @title Surface API
// @version 1.0
// @description API for managing hockey rink schedules, surfaces, locations, and events
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@example.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8000
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey CookieAuth
// @in cookie
// @name gosession

var db *gorm.DB
var cfg models.Config
var sess *session.Manager
var configFile = flag.String("config", "config.yaml", "Path to config file")

// rate limiting for password change: track failed attempts timestamps per username
var pwdChangeLock sync.Mutex
var pwdChangeAttempts = make(map[string][]time.Time)

const pwdChangeWindow = 5 * time.Minute
const pwdChangeMaxAttempts = 5

func init() {
	flag.Parse()
	gin.SetMode(gin.ReleaseMode)

	viper.SetConfigFile(*configFile)
	viper.SetDefault("port", "8000")
	viper.SetDefault("mode", "production")

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	err := viper.Unmarshal(&cfg)

	if err != nil {
		panic(err)
	}

	db, err = gorm.Open(mysql.Open(cfg.DB_DSN))
	if err != nil {
		panic(err)
	}
	log.Println(cfg)

	db.Exec(`SET SESSION sql_mode=(SELECT REPLACE(@@sql_mode,'ONLY_FULL_GROUP_BY',''))`)

	sess, err = session.NewManager("mysql", &session.ManagerConfig{
		CookieName:      "gosession",
		Gclifetime:      3600,
		ProviderConfig:  cfg.DB_DSN,
		EnableSetCookie: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	go sess.GC()
}

func main() {
	r := gin.Default()

	if cfg.Mode == "local" {
		corsCfg := cors.DefaultConfig()
		corsCfg.AllowCredentials = true
		corsCfg.AllowOrigins = []string{"http://localhost:5173"}
		r.Use(cors.New(corsCfg))
	}
	r.Use(AuthMiddleware)

	r.GET("/site-locations/:site", getSiteLoc)
	r.GET("/mhr-locations", getMHRLoc)
	r.POST("/mhr-set-location", setMHRLoc)
	r.POST("/mhr-set-surface", setMHRSurface)
	r.POST("/mhr-unset-mapping", unsetMHRMapping)
	r.GET("/mappings/:site", getMappings)
	r.GET("/sites", getSites)
	r.GET("/surfaces", getSurfaces)
	r.GET("/provinces", getProvinces)
	r.POST("/set-surface", setSurface)
	r.POST("/set-location", setLocation)
	r.POST("/set-mapping", setMapping)
	r.POST("/unset-mapping", unsetMapping)
	r.POST("/login", login)
	r.GET("/logout", logout)
	r.GET("/session", checkSession)
	r.GET("/report", surfaceReport)
	r.GET("/report/download", downloadReportCSV)
	r.GET("/rink-report", rinkReport)
	r.GET("/events-by-date", getEventsByDateRange)
	r.GET("/ramp-mappings/:province", rampMappings)
	r.GET("/ramp-provinces", rampProvinces)
	r.POST("/set-ramp-mapping", SetRampMappings)
	r.GET("/locations", getLocations)
	r.GET("/events", getEvents)
	r.PUT("/events/:id", updateEvent)

	// User management routes
	r.GET("/users", listUsers)
	r.POST("/users", addUser)
	r.DELETE("/users/:username", deleteUser)
	r.PUT("/users/:username/password", changePassword)

	// Sites config CRUD routes
	r.GET("/sites-config", getSitesConfig)
	r.GET("/parser-types", getParserTypes)
	r.GET("/sites-config/:id", getSitesConfigByID)
	r.POST("/sites-config", createSitesConfig)
	r.PUT("/sites-config/:id", updateSitesConfig)
	r.DELETE("/sites-config/:id", deleteSitesConfig)

	// Swagger documentation endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Println("starting server on ", cfg.Port)

	if err := r.Run(":" + cfg.Port); err != nil {
		panic(err)
	}
}

func surfaceReport(c *gin.Context) {
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
	if err := db.Raw(countQuery, args...).Scan(&total).Error; err != nil {
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
	if err := db.Raw(query, queryArgs...).Scan(&result).Error; err != nil {
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

func downloadReportCSV(c *gin.Context) {
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
	if err := db.Raw(query, args...).Scan(&result).Error; err != nil {
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

// @Summary Get rink usage report
// @Description Get paginated rink usage report aggregated by date and location. Supports filtering by rink (partial), province, city, and exporting CSV via 'export' query param.
// @Tags Reports
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param perPage query int false "Results per page" default(10)
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param rink query string false "Rink name (partial match)"
// @Param province query string false "Province ID"
// @Param city query string false "City name"
// @Param site query string false "Site name"
// @Param export query string false "If present returns CSV download"
// @Success 200 {object} RinkReportResponse
// @Failure 400 {object} map[string]interface{} "bad request"
// @Failure 500 {object} map[string]interface{} "error"
// @Security CookieAuth
// @Router /rink-report [get]
func rinkReport(c *gin.Context) {
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
	if err := db.Raw(countQuery, args...).Scan(&total).Error; err != nil {
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

	if err := db.Raw(query, queryArgs...).Scan(&result).Error; err != nil {
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

func getEventsByDateRange(c *gin.Context) {
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
	if err := db.First(&location, locationID).Error; err != nil {
		sendError(c, err)
		return
	}

	var results []models.EventWithLocation
	if err := db.Table("events").
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

func setSurface(c *gin.Context) {
	var input = &models.SiteLocResult{}

	if err := c.BindJSON(input); err != nil {
		sendError(c, err)
		return
	}

	if input.SurfaceID == -1 {
		if err := db.Model(input).Where("site=? and location=?", input.Site, input.Location).Select("SurfaceID").Updates(input).Error; err != nil {
			sendError(c, err)
			return
		}
	} else {
		var surface = &model.Surface{}
		if err := db.Find(surface, input.SurfaceID).Error; err != nil {
			sendError(c, err)
			return
		}

		input.LocationID = surface.LocationID

		if err := db.Model(input).Where("site=? and location=?", input.Site, input.Location).Select("LocationID", "SurfaceID").Updates(input).Error; err != nil {
			sendError(c, err)
			return
		}
	}
	var result = []models.SiteLocResult{}

	if err := db.Joins("LinkedSurface").Joins("LiveBarnLocation").Find(&result, "site=?", input.Site).Error; err != nil {
		sendError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// @Summary Set Location ID with a SiteLocation Record
// @Description Set Location ID with a SiteLocation Record
// @Tags Mappings
// @Accept json
// @Produce json
// @Param input body models.SetLocationInput true "Mapping to set"
// @Success 200 {array} models.SiteLocResult
// @Failure 500 {object} map[string]interface{} "error"
// @Security CookieAuth
// @Router /set-location [post]
func setLocation(c *gin.Context) {
	var input = struct {
		LocationId int32  `json:"location_id"`
		Site       string `json:"site"`
		Location   string `json:"location"`
	}{}

	if err := c.BindJSON(&input); err != nil {
		sendError(c, err)
		return
	}

	err := db.Exec(`update sites_locations set location_id=? where site=? and location=?`, input.LocationId, input.Site, input.Location).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error updating location"})
		return
	}

	c.AddParam("site", input.Site)
	getSiteLoc(c)
}

// @Summary Unset mapping
// @Description Remove location or surface mapping for a site location
// @Tags Mappings
// @Accept json
// @Produce json
// @Param input body models.UnsetMappingInput true "Mapping to unset"
// @Success 200 {array} models.SiteLocResult
// @Failure 400 {object} map[string]interface{} "error"
// @Failure 500 {object} map[string]interface{} "error"
// @Security CookieAuth
// @Router /unset-mapping [post]
func unsetMapping(c *gin.Context) {
	var input = &models.UnsetMappingInput{}

	if err := c.BindJSON(input); err != nil {
		sendError(c, err)
		return
	}

	updateFields := make(map[string]any)

	if input.Type == "location" {
		var current models.SiteLocResult
		if err := db.Model(&models.SiteLocResult{}).
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

	if err := db.Model(&models.SiteLocResult{}).
		Where("site = ? AND location = ?", input.Site, input.Location).
		Updates(updateFields).Error; err != nil {
		sendError(c, err)
		return
	}

	c.AddParam("site", input.Site)
	getSiteLoc(c)
}

func setMapping(c *gin.Context) {
	var input = &models.Mapping{}

	if err := c.BindJSON(input); err != nil {
		sendError(c, err)
		return
	}

	var table = input.Site + "_mappings"
	var surface = &model.Surface{}
	if err := db.Find(surface, input.SurfaceID).Error; err != nil {
		sendError(c, err)
		return
	}

	if err := db.Exec(fmt.Sprintf(`update %s set surface_id=? where location=?`, table), input.SurfaceID, input.Location).Error; err != nil {
		sendError(c, err)
		return
	}
	c.AddParam("site", input.Site)
	getMappings(c)
}

// @Summary Get surfaces
// @Description Get list of all surfaces with location details, optionally filtered by province
// @Tags Surfaces
// @Produce json
// @Param province query string false "Province name"
// @Success 200 {array} models.SurfaceResult
// @Failure 500 {object} map[string]interface{} "error"
// @Security CookieAuth
// @Router /surfaces [get]
func getSurfaces(c *gin.Context) {
	var surfaces = []models.SurfaceResult{}
	province := c.Query("province")

	var err error

	err = db.Raw(`SELECT
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

// @Summary Get provinces
// @Description Get list of all provinces in Canada
// @Tags Location
// @Produce json
// @Success 200 {array} object{id=int32,province_name=string}
// @Failure 500 {object} map[string]interface{} "error"
// @Security CookieAuth
// @Router /provinces [get]
func getProvinces(c *gin.Context) {
	var result = []struct {
		Id           int32  `json:"id"`
		ProvinceName string `json:"province_name"`
	}{}

	err := db.Raw(`select id, province_name from provinces order by province_name`).
		Scan(&result).Error

	if err != nil {
		sendError(c, err)
	}

	c.JSON(http.StatusOK, result)
}

// @Summary Get sites
// @Description Get list of all configured sites
// @Tags Sites
// @Produce json
// @Success 200 {array} models.SitesConfigResponse
// @Failure 500 {object} map[string]interface{} "error"
// @Security CookieAuth
// @Router /sites [get]
func getSites(c *gin.Context) {
	var sites = []model.SitesConfig{}

	if err := db.Order("display_name asc").Find(&sites).Error; err != nil {
		sendError(c, err)
		return
	}
	c.JSON(http.StatusOK, sites)
}

func getSiteLoc(c *gin.Context) {
	site := c.Param("site")
	var result = []models.SiteLocResult{}

	if err := db.Joins("LinkedSurface").Joins("LiveBarnLocation").Find(&result, "site=?", site).Error; err != nil {
		sendError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func getMHRLoc(c *gin.Context) {
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

	baseQuery := db.Model(&models.MhrLocation{})

	if name != "" {
		// baseQuery = baseQuery.Where(`rink_name like ?`, "%"+name+"%").Count(&total).Error; err != nil {
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

// @Summary Set Location ID with a MHRLocation Record
// @Description Set Location ID with a MHRLocation Record
// @Tags Mappings
// @Accept json
// @Produce json
// @Param input body models.SetLocationInput true "Mapping to set"
// @Success 200 {array} models.MHRLocResult
// @Failure 500 {object} map[string]interface{} "error"
// @Security CookieAuth
// @Router /mhr-set-location [post]
func setMHRLoc(c *gin.Context) {
	var input = struct {
		LocationId int32 `json:"location_id"`
		MhrId      int   `json:"mhr_id"`
	}{}

	if err := c.BindJSON(&input); err != nil {
		sendError(c, err)
		return
	}

	err := db.Exec(`update mhr_locations set livebarn_location_id=? where mhr_id=?`, input.LocationId, input.MhrId).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error updating location"})
		return
	}

	c.Status(http.StatusOK)
}

func setMHRSurface(c *gin.Context) {
	var input = &models.MhrLocation{}

	if err := c.BindJSON(input); err != nil {
		sendError(c, err)
		return
	}

	if input.LivebarnSurfaceId == -1 {
		if err := db.Exec(`update mhr_locations set livebarn_surface_id=? where mhr_id=?`, input.LivebarnSurfaceId, input.MhrID).Error; err != nil {
			sendError(c, err)
			return
		}
	} else {
		var surface = &model.Surface{}
		if err := db.Find(surface, input.LivebarnSurfaceId).Error; err != nil {
			sendError(c, err)
			return
		}

		input.LivebarnLocationId = int(surface.LocationID)

		if err := db.Exec(`update mhr_locations set livebarn_location_id=?, livebarn_surface_id=? where mhr_id=?`, input.LivebarnLocationId, input.LivebarnSurfaceId, input.MhrID).Error; err != nil {
			sendError(c, err)
			return
		}
	}
	c.Status(http.StatusOK)
}

// @Summary Unset MHR mapping
// @Description Remove location or surface mapping for a mhr location
// @Tags Mappings
// @Accept json
// @Produce json
// @Param input body models.UnsetMHRMappingInput true "Mapping to unset"
// @Success 200
// @Failure 400 {object} map[string]interface{} "error"
// @Failure 500 {object} map[string]interface{} "error"
// @Security CookieAuth
// @Router /mhr-unset-mapping [post]
func unsetMHRMapping(c *gin.Context) {
	var input = &models.UnsetMHRMappingInput{}

	if err := c.BindJSON(input); err != nil {
		sendError(c, err)
		return
	}

	updateFields := make(map[string]any)

	if input.Type == "location" {
		var current models.MhrLocation
		if err := db.Model(&models.MhrLocation{}).
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

	if err := db.Model(&models.MhrLocation{}).
		Where("mhr_id = ?", input.MhrId).
		Updates(updateFields).Error; err != nil {
		sendError(c, err)
		return
	}
	c.Status(http.StatusOK)
}

func getMappings(c *gin.Context) {
	site := c.Param("site")
	var table = site + "_mappings"
	var result = []models.Mapping{}

	err := db.Raw(fmt.Sprintf(`SELECT
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

func sendError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": err.Error(),
	})
}

func login(c *gin.Context) {
	var req = &models.Login{}

	if err := c.BindJSON(req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	hash := sha256.Sum256([]byte(req.Password))
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(hash)))
	base64.StdEncoding.Encode(dst, hash[:])

	if err := db.First(req, "username=?", req.Username).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{
				"error": "Invalid username/password",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if req.Password != string(dst) {
		c.JSON(http.StatusOK, gin.H{
			"error": "Invalid username/password",
		})
		return
	}

	s, _ := c.Get("sess")
	sess := s.(session.Store)
	sess.Set("username", req.Username)

	c.JSON(http.StatusOK, gin.H{
		"username": req.Username,
	})
}

func AuthMiddleware(c *gin.Context) {
	s, err := sess.SessionStart(c.Writer, c.Request)
	if err != nil {
		log.Println("session error", err)
	}
	defer s.SessionRelease(c.Writer)

	url := c.Request.URL.String()
	if url != "/unset-mapping" && url != "/swagger/" && url != "/login" && url != "/logout" {
		if s.Get("username") == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Session expired",
			})
			return
		}
	}
	c.Set("sess", s)
	c.Next()
}

// @Summary Check session
// @Description Get current session information
// @Tags Authentication
// @Produce json
// @Success 200 {object} map[string]interface{} "username"
// @Security CookieAuth
// @Router /session [get]
func checkSession(c *gin.Context) {
	s, _ := c.Get("sess")
	sess := s.(session.Store)
	username := sess.Get("username")

	c.JSON(http.StatusOK, gin.H{
		"username": username,
	})
}

// @Summary Logout
// @Description Destroy current session
// @Tags Authentication
// @Success 200 "OK"
// @Security CookieAuth
// @Router /logout [get]
func logout(c *gin.Context) {
	sess.SessionDestroy(c.Writer, c.Request)
	c.Status(http.StatusOK)
}

func rampMappings(c *gin.Context) {
	var province = c.Param("province")
	var result []models.RampLocation

	err := db.Raw(`SELECT
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

func rampProvinces(c *gin.Context) {
	var result []sql.NullString

	err := db.Raw("select distinct(province_name) provinces from RAMP_Locations order by provinces").Scan(&result).Error

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

func SetRampMappings(c *gin.Context) {
	var input = &models.SetRampSurfaceID{}

	if err := c.BindJSON(input); err != nil {
		sendError(c, err)
		return
	}

	var rec = &models.RampLocation{}

	if err := db.First(rec, "rarid=?", input.RarID).Error; err != nil {
		sendError(c, err)
		return
	}

	if err := db.Exec(`UPDATE RAMP_Locations SET surface_id=? where rarid=?`, input.SurfaceID, input.RarID).Error; err != nil {
		sendError(c, err)
		return
	}

	c.AddParam("province", input.Province)
	rampMappings(c)
}

// getSitesConfig retrieves all sites config records
func getSitesConfig(c *gin.Context) {
	var sitesConfigs []model.SitesConfig

	if err := db.Find(&sitesConfigs).Error; err != nil {
		sendError(c, err)
		return
	}

	var response []models.SitesConfigResponse
	for _, sc := range sitesConfigs {
		response = append(response, convertToSitesConfigResponse(sc))
	}

	c.JSON(http.StatusOK, response)
}

// getSitesConfigByID retrieves a single sites config record by ID
func getSitesConfigByID(c *gin.Context) {
	id := c.Param("id")
	var sitesConfig model.SitesConfig

	if err := db.First(&sitesConfig, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Sites config not found"})
			return
		}
		sendError(c, err)
		return
	}

	c.JSON(http.StatusOK, convertToSitesConfigResponse(sitesConfig))
}

// createSitesConfig creates a new sites config record
func createSitesConfig(c *gin.Context) {
	var input models.SitesConfigInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sitesConfig := convertToSitesConfigModel(input)

	if err := db.Create(&sitesConfig).Error; err != nil {
		sendError(c, err)
		return
	}

	c.JSON(http.StatusCreated, convertToSitesConfigResponse(sitesConfig))
}

// updateSitesConfig updates an existing sites config record
func updateSitesConfig(c *gin.Context) {
	id := c.Param("id")
	var input models.SitesConfigInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var sitesConfig model.SitesConfig
	if err := db.First(&sitesConfig, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Sites config not found"})
			return
		}
		sendError(c, err)
		return
	}

	updatedConfig := convertToSitesConfigModel(input)
	updatedConfig.ID = sitesConfig.ID
	updatedConfig.CreatedAt = sitesConfig.CreatedAt

	if err := db.Save(&updatedConfig).Error; err != nil {
		sendError(c, err)
		return
	}

	c.JSON(http.StatusOK, convertToSitesConfigResponse(updatedConfig))
}

// deleteSitesConfig deletes a sites config record
func deleteSitesConfig(c *gin.Context) {
	id := c.Param("id")
	var sitesConfig model.SitesConfig

	if err := db.First(&sitesConfig, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Sites config not found"})
			return
		}
		sendError(c, err)
		return
	}

	if err := db.Delete(&sitesConfig).Error; err != nil {
		sendError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sites config deleted successfully"})
}

// Helper function to convert model to response
func convertToSitesConfigResponse(sc model.SitesConfig) models.SitesConfigResponse {
	response := models.SitesConfigResponse{
		ID:         sc.ID,
		SiteName:   sc.SiteName,
		BaseURL:    sc.BaseURL,
		ParserType: sc.ParserType,
		CreatedAt:  sc.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  sc.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if sc.DisplayName.Valid {
		response.DisplayName = &sc.DisplayName.String
	}
	if sc.HomeTeam.Valid {
		response.HomeTeam = &sc.HomeTeam.String
	}
	if sc.ParserConfig != nil {
		response.ParserConfig = *sc.ParserConfig
	}
	if sc.Enabled.Valid {
		response.Enabled = &sc.Enabled.Bool
	}
	if sc.LastScrapedAt.Valid {
		lastScraped := sc.LastScrapedAt.Time.Format("2006-01-02 15:04:05")
		response.LastScrapedAt = &lastScraped
	}
	if sc.ScrapeFrequencyHours.Valid {
		response.ScrapeFrequencyHours = &sc.ScrapeFrequencyHours.Int32
	}
	if sc.Notes.Valid {
		response.Notes = &sc.Notes.String
	}

	return response
}

// Helper function to convert input to model
func convertToSitesConfigModel(input models.SitesConfigInput) model.SitesConfig {
	sitesConfig := model.SitesConfig{
		SiteName:   input.SiteName,
		BaseURL:    input.BaseURL,
		ParserType: input.ParserType,
	}

	if input.DisplayName != nil {
		sitesConfig.DisplayName = sql.NullString{String: *input.DisplayName, Valid: true}
	}
	if input.HomeTeam != nil {
		sitesConfig.HomeTeam = sql.NullString{String: *input.HomeTeam, Valid: true}
	}
	if input.ParserConfig != nil {
		pc := model.ParserConfig(input.ParserConfig)
		sitesConfig.ParserConfig = &pc
	}
	if input.Enabled != nil {
		sitesConfig.Enabled = sql.NullBool{Bool: *input.Enabled, Valid: true}
	}
	if input.ScrapeFrequencyHours != nil {
		sitesConfig.ScrapeFrequencyHours = sql.NullInt32{Int32: *input.ScrapeFrequencyHours, Valid: true}
	}
	if input.Notes != nil {
		sitesConfig.Notes = sql.NullString{String: *input.Notes, Valid: true}
	}

	return sitesConfig
}

// getParserTypes returns all available parser types
func getParserTypes(c *gin.Context) {
	parserTypes := []string{
		"day_details",
		"day_details_parser1",
		"day_details_parser2",
		"month_based",
		"group_based",
		"custom",
		"external",
	}

	c.JSON(http.StatusOK, parserTypes)
}

// @Summary Get locations
// @Description Get paginated list of locations with their surfaces
// @Tags Locations
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param perPage query int false "Results per page" default(10)
// @Param name query string false "Filter by location name"
// @Param postal_code query string false "Filter by postal code"
// @Success 200 {object} map[string]interface{} "data, page, perPage, total"
// @Failure 500 {object} map[string]interface{} "error"
// @Security CookieAuth
// @Router /locations [get]
// LocationWithSurfaces represents a location with its associated surfaces
type LocationWithSurfaces struct {
	model.Location
	Surfaces []model.Surface `json:"surfaces" gorm:"foreignKey:LocationID;references:ID"`
}

// getLocations returns all locations with their associated surfaces
// @Summary Get locations
// @Description Get paginated list of locations with their surfaces
// @Tags Locations
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param perPage query int false "Results per page" default(10)
// @Param name query string false "Filter by location name"
// @Param postal_code query string false "Filter by postal code"
// @Success 200 {object} map[string]interface{} "data, page, perPage, total"
// @Failure 500 {object} map[string]interface{} "error"
// @Security CookieAuth
// @Router /locations [get]
func getLocations(c *gin.Context) {
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

	baseQuery := db.Model(&model.Location{}).Order("locations.name")

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

// @Summary Get events
// @Description Get paginated list of events with optional filters
// @Tags Events
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param perPage query int false "Results per page" default(10)
// @Param site query string false "Filter by site"
// @Param surface_id query int false "Filter by surface ID"
// @Param start_date query string false "Filter by start date (YYYY-MM-DD)"
// @Param end_date query string false "Filter by end date (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{} "data, page, perPage, total"
// @Failure 500 {object} map[string]interface{} "error"
// @Security CookieAuth
// @Router /events [get]
// getEvents returns all events with pagination
func getEvents(c *gin.Context) {
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

	baseQuery := db.Model(&model.Event{}).Order("datetime DESC")

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
		c.Writer.Header().Add("content-disposition", "attachment;filename=rink_report.csv")
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

type UpdateEventInput struct {
	SurfaceID    int32 `json:"surface_id"`
	UpdateFuture bool  `json:"update_future"`
}

// updateEvent updates the surface_id and location_id of an event
func updateEvent(c *gin.Context) {
	id := c.Param("id")
	var input UpdateEventInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var event model.Event
	if err := db.First(&event, id).Error; err != nil {
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
		if err := db.First(&surface, input.SurfaceID).Error; err != nil {
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
		result := db.Model(&model.Event{}).
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

		if err := db.Save(&event).Error; err != nil {
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

// @Summary List users
// @Description Get list of all users (passwords are masked)
// @Tags Users
// @Produce json
// @Success 200 {array} models.Login
// @Failure 500 {object} map[string]interface{} "error"
// @Security CookieAuth
// @Router /users [get]
func listUsers(c *gin.Context) {
	var users []models.Login

	if err := db.Find(&users).Error; err != nil {
		sendError(c, err)
		return
	}

	for i := range users {
		users[i].Password = ""
	}

	c.JSON(http.StatusOK, users)
}

// @Summary Add user
// @Description Create a new user
// @Tags Users
// @Accept json
// @Produce json
// @Param user body models.CreateUserInput true "User information"
// @Success 201 {object} models.Login
// @Failure 400 {object} map[string]interface{} "error"
// @Failure 500 {object} map[string]interface{} "error"
// @Security CookieAuth
// @Router /users [post]
func addUser(c *gin.Context) {
	var input models.CreateUserInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash := sha256.Sum256([]byte(input.Password))
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(hash)))
	base64.StdEncoding.Encode(dst, hash[:])

	user := models.Login{
		Username: input.Username,
		Password: string(dst),
	}

	if err := db.Create(&user).Error; err != nil {
		sendError(c, err)
		return
	}

	user.Password = ""
	c.JSON(http.StatusCreated, user)
}

// @Summary Delete user
// @Description Delete a user and invalidate their sessions (cannot delete currently logged in user)
// @Tags Users
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} map[string]interface{} "message"
// @Failure 400 {object} map[string]interface{} "error"
// @Failure 404 {object} map[string]interface{} "error"
// @Failure 500 {object} map[string]interface{} "error"
// @Security CookieAuth
// @Router /users/{username} [delete]
func deleteUser(c *gin.Context) {
	username := c.Param("username")

	// Get current logged in user
	s, _ := c.Get("sess")
	sess := s.(session.Store)
	currentUser := sess.Get("username")

	// Prevent self-deletion
	if currentUser != nil && currentUser.(string) == username {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete currently logged in user"})
		return
	}

	var user models.Login

	if err := db.First(&user, "username = ?", username).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		sendError(c, err)
		return
	}

	usernameToDelete := user.Username

	if err := db.Delete(&user).Error; err != nil {
		sendError(c, err)
		return
	}

	go invalidateUserSessions(usernameToDelete)

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// @Summary Change own password
// @Description Allow a logged-in user to change their own password. Accepts {"current_password":"old","password":"new","confirm":"new"}.
// @Tags Users
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Param input body object true "New password"
// @Success 200 {object} map[string]interface{} "Password changed successfully. Server invalidates sessions; client should force re-login. Example: {\"message\":\"password changed\"}"
// @Failure 400 {object} map[string]interface{} "Bad Request. Returned when validation fails or current password is incorrect. Response shape: {\"error\":\"message\", optionally \"attempts_left\":int, \"cooldown_seconds\":int}"
// @Failure 429 {object} map[string]interface{} "Too Many Requests. Returned when rate limit reached. Response shape: {\"error\":\"too many password change attempts, try again later\", \"attempts_left\":0, \"cooldown_seconds\":int}"
// @Failure 500 {object} map[string]interface{} "Internal Server Error: {\"error\":\"message\"}"
// @Security CookieAuth
// @Router /users/{username}/password [put]
func changePassword(c *gin.Context) {
	username := c.Param("username")

	s, _ := c.Get("sess")
	sess := s.(session.Store)
	currentUser := sess.Get("username")

	// Only allow changing own password
	if currentUser == nil || currentUser.(string) != username {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can only change own password"})
		return
	}

	// rate limiting: block if too many recent failed attempts
	now := time.Now()
	pwdChangeLock.Lock()
	attempts := pwdChangeAttempts[username]
	var pruned []time.Time
	for _, t := range attempts {
		if now.Sub(t) < pwdChangeWindow {
			pruned = append(pruned, t)
		}
	}
	if len(pruned) >= pwdChangeMaxAttempts {
		// compute remaining cooldown based on the oldest attempt in window
		earliest := pruned[0]
		remaining := max(0, pwdChangeWindow-now.Sub(earliest))
		pwdChangeLock.Unlock()
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many password change attempts, try again later", "attempts_left": 0, "cooldown_seconds": int(remaining.Seconds())})
		return
	}
	pwdChangeAttempts[username] = pruned
	pwdChangeLock.Unlock()

	var input struct {
		CurrentPassword string `json:"current_password"`
		Password        string `json:"password"`
		Confirm         string `json:"confirm"`
	}

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// basic validations
	if input.Password == "" || input.Confirm == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password and confirm are required"})
		return
	}
	if input.Password != input.Confirm {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password and confirm do not match"})
		return
	}
	if len(input.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 8 characters"})
		return
	}

	// verify current password
	var user models.Login
	if err := db.First(&user, "username = ?", username).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
			return
		}
		sendError(c, err)
		return
	}

	currHash := sha256.Sum256([]byte(input.CurrentPassword))
	dstCurr := make([]byte, base64.StdEncoding.EncodedLen(len(currHash)))
	base64.StdEncoding.Encode(dstCurr, currHash[:])

	if user.Password != string(dstCurr) {
		// record failed attempt
		pwdChangeLock.Lock()
		pwdChangeAttempts[username] = append(pwdChangeAttempts[username], now)
		// compute attempts left and cooldown
		attemptsLeft := pwdChangeMaxAttempts - len(pwdChangeAttempts[username])
		attemptsLeft = max(0, attemptsLeft)

		var cooldownSeconds int
		if attemptsLeft == 0 {
			earliest := pwdChangeAttempts[username][0]
			remaining := pwdChangeWindow - now.Sub(earliest)
			remaining = max(0, remaining)

			cooldownSeconds = int(remaining.Seconds())
		} else {
			cooldownSeconds = 0
		}
		pwdChangeLock.Unlock()

		c.JSON(http.StatusBadRequest, gin.H{"error": "current password is incorrect", "attempts_left": attemptsLeft, "cooldown_seconds": cooldownSeconds})
		return
	}

	// set new password
	newHash := sha256.Sum256([]byte(input.Password))
	dstNew := make([]byte, base64.StdEncoding.EncodedLen(len(newHash)))
	base64.StdEncoding.Encode(dstNew, newHash[:])

	if err := db.Model(&models.Login{}).Where("username = ?", username).Update("password", string(dstNew)).Error; err != nil {
		sendError(c, err)
		return
	}

	// invalidate other sessions in background and ensure current session remains set
	go invalidateUserSessions(username)
	// clear failed attempts on success
	pwdChangeLock.Lock()
	delete(pwdChangeAttempts, username)
	pwdChangeLock.Unlock()

	// UI should force re-login sess.Set("username", username)

	c.Status(http.StatusOK)
}

func invalidateUserSessions(username string) {
	type sessionRecord struct {
		SessionKey  string `gorm:"column:session_key"`
		SessionData []byte `gorm:"column:session_data"`
	}

	var sessions []sessionRecord
	if err := db.Table("session").Find(&sessions).Error; err != nil {
		log.Println("Warning: Failed to fetch sessions:", err)
		return
	}

	var keysToDelete []string
	for _, session := range sessions {
		sessionDataStr := string(session.SessionData)
		if len(sessionDataStr) > 0 && (string(session.SessionData)[0] == 0x0D || session.SessionData[0] == 0x00) {
			if bytes.Contains(session.SessionData, []byte(username)) {
				keysToDelete = append(keysToDelete, session.SessionKey)
			}
		}
	}

	if len(keysToDelete) > 0 {
		if err := db.Table("session").Where("session_key IN ?", keysToDelete).Delete(nil).Error; err != nil {
			log.Println("Warning: Failed to delete sessions:", err)
		} else {
			log.Printf("Invalidated %d session(s) for user: %s", len(keysToDelete), username)
		}
	}
}
