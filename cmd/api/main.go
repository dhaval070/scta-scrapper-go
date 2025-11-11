package main

import (
	"calendar-scrapper/config"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type SiteConfig struct {
	ID                   int         `json:"id"`
	SiteName             string      `json:"site_name"`
	DisplayName          *string     `json:"display_name"`
	BaseURL              string      `json:"base_url"`
	HomeTeam             *string     `json:"home_team"`
	ParserType           string      `json:"parser_type"`
	ParserConfig         interface{} `json:"parser_config"`
	Enabled              bool        `json:"enabled"`
	LastScrapedAt        *string     `json:"last_scraped_at"`
	ScrapeFrequencyHours int         `json:"scrape_frequency_hours"`
	Notes                *string     `json:"notes"`
	CreatedAt            string      `json:"created_at"`
	UpdatedAt            string      `json:"updated_at"`
}

var db *sql.DB

func main() {
	// Initialize config
	config.Init("config", ".", "..")
	cfg := config.MustReadConfig()

	// Get database connection from config or environment or use default
	dsn := cfg.DbDSN

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Cannot connect to database:", err)
	}

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// API routes
	e.GET("/api/sites", listSites)
	e.GET("/api/sites/:id", getSite)
	e.PUT("/api/sites/:id", updateSite)
	e.POST("/api/sites", createSite)
	e.DELETE("/api/sites/:id", deleteSite)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("API server starting on port %s\n", port)
	e.Logger.Fatal(e.Start(":" + port))
}

func listSites(c echo.Context) error {
	rows, err := db.Query(`
		SELECT id, site_name, display_name, base_url, home_team, parser_type, 
		       parser_config, enabled, last_scraped_at, scrape_frequency_hours, 
		       notes, created_at, updated_at
		FROM sites_config
		ORDER BY site_name
	`)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	defer rows.Close()

	var sites []SiteConfig
	for rows.Next() {
		var site SiteConfig
		var parserConfigJSON sql.NullString
		var displayName, homeTeam, notes, lastScrapedAt sql.NullString
		err := rows.Scan(
			&site.ID, &site.SiteName, &displayName, &site.BaseURL,
			&homeTeam, &site.ParserType, &parserConfigJSON, &site.Enabled,
			&lastScrapedAt, &site.ScrapeFrequencyHours, &notes,
			&site.CreatedAt, &site.UpdatedAt,
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		if displayName.Valid {
			site.DisplayName = &displayName.String
		}
		if homeTeam.Valid {
			site.HomeTeam = &homeTeam.String
		}
		if notes.Valid {
			site.Notes = &notes.String
		}
		if lastScrapedAt.Valid {
			site.LastScrapedAt = &lastScrapedAt.String
		}

		if parserConfigJSON.Valid {
			var parserConfig interface{}
			if err := json.Unmarshal([]byte(parserConfigJSON.String), &parserConfig); err == nil {
				site.ParserConfig = parserConfig
			}
		}

		sites = append(sites, site)
	}

	return c.JSON(http.StatusOK, sites)
}

func getSite(c echo.Context) error {
	id := c.Param("id")

	var site SiteConfig
	var parserConfigJSON sql.NullString
	var displayName, homeTeam, notes, lastScrapedAt sql.NullString
	err := db.QueryRow(`
		SELECT id, site_name, display_name, base_url, home_team, parser_type, 
		       parser_config, enabled, last_scraped_at, scrape_frequency_hours, 
		       notes, created_at, updated_at
		FROM sites_config
		WHERE id = ?
	`, id).Scan(
		&site.ID, &site.SiteName, &displayName, &site.BaseURL,
		&homeTeam, &site.ParserType, &parserConfigJSON, &site.Enabled,
		&lastScrapedAt, &site.ScrapeFrequencyHours, &notes,
		&site.CreatedAt, &site.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Site not found"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if displayName.Valid {
		site.DisplayName = &displayName.String
	}
	if homeTeam.Valid {
		site.HomeTeam = &homeTeam.String
	}
	if notes.Valid {
		site.Notes = &notes.String
	}
	if lastScrapedAt.Valid {
		site.LastScrapedAt = &lastScrapedAt.String
	}

	if parserConfigJSON.Valid {
		var parserConfig interface{}
		if err := json.Unmarshal([]byte(parserConfigJSON.String), &parserConfig); err == nil {
			site.ParserConfig = parserConfig
		}
	}

	return c.JSON(http.StatusOK, site)
}

func updateSite(c echo.Context) error {
	id := c.Param("id")

	var site SiteConfig
	if err := c.Bind(&site); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	var parserConfigJSON sql.NullString
	if site.ParserConfig != nil {
		jsonBytes, err := json.Marshal(site.ParserConfig)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid parser config"})
		}
		parserConfigJSON = sql.NullString{String: string(jsonBytes), Valid: true}
	}

	var displayName, homeTeam, notes sql.NullString
	if site.DisplayName != nil {
		displayName = sql.NullString{String: *site.DisplayName, Valid: true}
	}
	if site.HomeTeam != nil {
		homeTeam = sql.NullString{String: *site.HomeTeam, Valid: true}
	}
	if site.Notes != nil {
		notes = sql.NullString{String: *site.Notes, Valid: true}
	}

	_, err := db.Exec(`
		UPDATE sites_config 
		SET site_name = ?, display_name = ?, base_url = ?, home_team = ?,
		    parser_type = ?, parser_config = ?, enabled = ?, 
		    scrape_frequency_hours = ?, notes = ?
		WHERE id = ?
	`, site.SiteName, displayName, site.BaseURL, homeTeam,
		site.ParserType, parserConfigJSON, site.Enabled,
		site.ScrapeFrequencyHours, notes, id)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Site updated successfully"})
}

func createSite(c echo.Context) error {
	var site SiteConfig
	if err := c.Bind(&site); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	var parserConfigJSON sql.NullString
	if site.ParserConfig != nil {
		jsonBytes, err := json.Marshal(site.ParserConfig)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid parser config"})
		}
		parserConfigJSON = sql.NullString{String: string(jsonBytes), Valid: true}
	}

	var displayName, homeTeam, notes sql.NullString
	if site.DisplayName != nil {
		displayName = sql.NullString{String: *site.DisplayName, Valid: true}
	}
	if site.HomeTeam != nil {
		homeTeam = sql.NullString{String: *site.HomeTeam, Valid: true}
	}
	if site.Notes != nil {
		notes = sql.NullString{String: *site.Notes, Valid: true}
	}

	result, err := db.Exec(`
		INSERT INTO sites_config 
		(site_name, display_name, base_url, home_team, parser_type, 
		 parser_config, enabled, scrape_frequency_hours, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, site.SiteName, displayName, site.BaseURL, homeTeam,
		site.ParserType, parserConfigJSON, site.Enabled,
		site.ScrapeFrequencyHours, notes)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	lastID, _ := result.LastInsertId()
	site.ID = int(lastID)

	return c.JSON(http.StatusCreated, site)
}

func deleteSite(c echo.Context) error {
	id := c.Param("id")

	_, err := db.Exec("DELETE FROM sites_config WHERE id = ?", id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Site deleted successfully"})
}
