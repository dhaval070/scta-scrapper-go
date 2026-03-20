package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Scraping status constants
const (
	StatusIdle      = "idle"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
)

// ScrapeRequest represents a request to trigger scraping
type ScrapeRequest struct {
	Site    string `json:"site"`    // Site name (empty for all enabled sites)
	Date    string `json:"date"`    // Date in YYYY-MM-DD format (optional)
	Workers int    `json:"workers"` // Number of workers (optional, default 1)
}

// ScrapeStatus represents current scraping status for a site
type ScrapeStatus struct {
	Site             string     `json:"site"`
	Status           string     `json:"status"`              // idle, running, completed, failed
	StartedAt        *time.Time `json:"started_at"`          // When scraping started
	Error            *string    `json:"error"`               // Error message if failed
	LastScrapedAt    *time.Time `json:"last_scraped_at"`     // From sites_config
	ScrapeFrequency  int        `json:"scrape_frequency"`    // Frequency in hours
	IsDueForScraping bool       `json:"is_due_for_scraping"` // Whether site is due for scraping
}

// triggerScrape starts scraping for the specified site(s)
// @Summary Trigger scraping for site(s)
// @Description Starts scraping for specified site or all enabled sites. Sites already being scraped are skipped. Returns 409 only if ALL requested sites are already running.
// @Tags Scraping
// @Accept json
// @Produce json
// @Param request body ScrapeRequest true "Scraping request parameters"
// @Success 202 {object} map[string]interface{} "Accepted, returns job status with sites_started and optionally sites_skipped"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 409 {object} map[string]interface{} "All requested sites are already being scraped"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security CookieAuth
// @Router /scrape [post]
func (app *App) triggerScrape(c *gin.Context) {
	var req ScrapeRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Determine sites to scrape
	var sites []string
	if req.Site == "" {
		// Get all enabled sites
		var enabledSites []struct {
			SiteName string
		}
		if err := app.db.Table("sites_config").Where("enabled = ?", true).Select("site_name").Find(&enabledSites).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch enabled sites: %v", err)})
			return
		}
		for _, s := range enabledSites {
			sites = append(sites, s.SiteName)
		}
		if len(sites) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No enabled sites found"})
			return
		}
	} else {
		sites = []string{req.Site}
	}

	// Filter out sites that are already being scraped
	app.scrapingMu.Lock()
	defer app.scrapingMu.Unlock()

	sitesToScrape := []string{}
	alreadyRunningSites := []string{}

	for _, site := range sites {
		isRunning := false

		// Check in-memory map first
		if _, exists := app.scrapingProcesses[site]; exists {
			isRunning = true
		}

		// Also check database status
		var status string
		if err := app.db.Table("sites_config").Where("site_name = ?", site).Select("scraping_status").Scan(&status).Error; err != nil {
			// Continue anyway
		}
		if status == StatusRunning {
			isRunning = true
		}

		if isRunning {
			alreadyRunningSites = append(alreadyRunningSites, site)
		} else {
			sitesToScrape = append(sitesToScrape, site)
		}
	}

	// If no sites remain to scrape
	if len(sitesToScrape) == 0 {
		if len(alreadyRunningSites) > 0 {
			c.JSON(http.StatusConflict, gin.H{
				"error":           "All requested sites are already being scraped",
				"already_running": alreadyRunningSites,
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No sites to scrape"})
		}
		return
	}

	// Update database status for each site that will be scraped
	for _, site := range sitesToScrape {
		if err := app.db.Exec(`
			UPDATE sites_config 
			SET scraping_status = ?, 
				scraping_started_at = NOW(),
				scraping_error = NULL
			WHERE site_name = ?
		`, StatusRunning, site).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update status for %s: %v", site, err)})
			return
		}
	}

	// Prepare scraper command - always use --sites with comma-separated list
	cmdArgs := []string{}
	// Join sites with commas for the --sites flag
	sitesList := strings.Join(sitesToScrape, ",")
	cmdArgs = append(cmdArgs, "--sites", sitesList)

	// Add import-locations flag (always import locations)
	cmdArgs = append(cmdArgs, "--import-locations")

	// Add date if provided
	if req.Date != "" {
		// Convert YYYY-MM-DD to MMyyyy
		parts := strings.Split(req.Date, "-")
		if len(parts) == 3 {
			mmyyyy := parts[1] + parts[0] // MMYYYY
			cmdArgs = append(cmdArgs, "--date", mmyyyy)
		} else {
			// Use as-is (might already be MMyyyy)
			cmdArgs = append(cmdArgs, "--date", req.Date)
		}
	} else {
		// Use current month
		now := time.Now()
		mmyyyy := fmt.Sprintf("%02d%d", now.Month(), now.Year())
		cmdArgs = append(cmdArgs, "--date", mmyyyy)
	}

	// Add workers if specified
	if req.Workers > 0 {
		cmdArgs = append(cmdArgs, "--workers", fmt.Sprintf("%d", req.Workers))
	}

	// Use configured scraper binary path
	if app.cfg.ScraperPath == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Scraper path not configured. Set scraper_path in config.yaml"})
		return
	}
	scraperPath := app.cfg.ScraperPath
	// Check if scraper binary exists and is executable
	if _, err := os.Stat(scraperPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Scraper binary not found at %s: %v", scraperPath, err)})
		return
	}
	projectRoot := filepath.Dir(scraperPath)

	// Create command
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, scraperPath, cmdArgs...)
	cmd.Dir = projectRoot

	// Capture stdout and stderr
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create stdout pipe: %v", err)})
		return
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create stderr pipe: %v", err)})
		return
	}

	// Store command in map
	for _, site := range sitesToScrape {
		app.scrapingProcesses[site] = cmd
	}

	// Start scraping in goroutine
	go func() {
		defer func() {
			app.scrapingMu.Lock()
			for _, site := range sitesToScrape {
				delete(app.scrapingProcesses, site)
			}
			app.scrapingMu.Unlock()
			cancel()
		}()

		log.Printf("Starting scraper for sites: %v", sitesToScrape)

		// Start command
		if err := cmd.Start(); err != nil {
			app.scrapingMu.Lock()
			defer app.scrapingMu.Unlock()
			app.updateScrapingStatus(sitesToScrape, StatusFailed, err.Error())
			return
		}

		// Read output concurrently
		stdoutCh := readPipeAsync(stdoutPipe)
		stderrCh := readPipeAsync(stderrPipe)

		// Wait for command to complete
		err := cmd.Wait()

		// Collect output
		stdout := <-stdoutCh
		stderr := <-stderrCh

		app.scrapingMu.Lock()
		defer app.scrapingMu.Unlock()

		// Update database status
		status := StatusCompleted
		errorMsg := ""
		if err != nil {
			status = StatusFailed
			// Combine error message with stderr
			if stderr != "" {
				errorMsg = fmt.Sprintf("%v\nSTDERR:\n%s", err, stderr)
			} else {
				errorMsg = err.Error()
			}
			log.Printf("Scraper failed for sites %v: %v\nSTDOUT:\n%s\nSTDERR:\n%s", sitesToScrape, err, stdout, stderr)
		} else {
			log.Printf("Scraper completed successfully for sites %v", sitesToScrape)
			// Log output for debugging
			if stdout != "" || stderr != "" {
				log.Printf("Scraper output for sites %v:\nSTDOUT:\n%s\nSTDERR:\n%s", sitesToScrape, stdout, stderr)
			}
		}

		app.updateScrapingStatus(sitesToScrape, status, errorMsg)
	}()

	response := gin.H{
		"message":       "Scraping started",
		"sites_started": sitesToScrape,
	}
	if len(alreadyRunningSites) > 0 {
		response["sites_skipped"] = alreadyRunningSites
		response["warning"] = fmt.Sprintf("%d site(s) were already running", len(alreadyRunningSites))
	}

	c.JSON(http.StatusAccepted, response)
}

// getScrapeStatus returns scraping status for a specific site
// @Summary Get scraping status for a site
// @Description Returns current scraping status for the specified site
// @Tags Scraping
// @Accept json
// @Produce json
// @Param site path string true "Site name"
// @Success 200 {object} ScrapeStatus
// @Failure 404 {object} map[string]string "Site not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security CookieAuth
// @Router /scrape/status/{site} [get]
func (app *App) getScrapeStatus(c *gin.Context) {
	site := c.Param("site")

	var status struct {
		SiteName             string     `gorm:"column:site_name"`
		ScrapingStatus       string     `gorm:"column:scraping_status"`
		ScrapingStartedAt    *time.Time `gorm:"column:scraping_started_at"`
		ScrapingError        *string    `gorm:"column:scraping_error"`
		LastScrapedAt        *time.Time `gorm:"column:last_scraped_at"`
		ScrapeFrequencyHours int        `gorm:"column:scrape_frequency_hours"`
	}

	err := app.db.Table("sites_config").
		Select("site_name, scraping_status, scraping_started_at, scraping_error, last_scraped_at, scrape_frequency_hours").
		Where("site_name = ?", site).
		First(&status).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Site not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if site is due for scraping
	isDue := false
	if status.LastScrapedAt == nil {
		isDue = true
	} else {
		dueTime := status.LastScrapedAt.Add(time.Duration(status.ScrapeFrequencyHours) * time.Hour)
		isDue = time.Now().After(dueTime)
	}

	// Check if site is currently being scraped in this process
	app.scrapingMu.Lock()
	_, isRunning := app.scrapingProcesses[site]
	app.scrapingMu.Unlock()

	// If database says running but not in our map, it might be stale
	finalStatus := status.ScrapingStatus
	if finalStatus == StatusRunning && !isRunning {
		// Stale status, reset to idle
		finalStatus = StatusIdle
	}

	c.JSON(http.StatusOK, ScrapeStatus{
		Site:             status.SiteName,
		Status:           finalStatus,
		StartedAt:        status.ScrapingStartedAt,
		Error:            status.ScrapingError,
		LastScrapedAt:    status.LastScrapedAt,
		ScrapeFrequency:  status.ScrapeFrequencyHours,
		IsDueForScraping: isDue,
	})
}

// getAllScrapeStatus returns scraping status for all sites
// @Summary Get scraping status for all sites
// @Description Returns current scraping status for all configured sites
// @Tags Scraping
// @Accept json
// @Produce json
// @Success 200 {array} ScrapeStatus
// @Failure 500 {object} map[string]string "Internal server error"
// @Security CookieAuth
// @Router /scrape/status [get]
func (app *App) getAllScrapeStatus(c *gin.Context) {
	var statuses []struct {
		SiteName             string     `gorm:"column:site_name"`
		ScrapingStatus       string     `gorm:"column:scraping_status"`
		ScrapingStartedAt    *time.Time `gorm:"column:scraping_started_at"`
		ScrapingError        *string    `gorm:"column:scraping_error"`
		LastScrapedAt        *time.Time `gorm:"column:last_scraped_at"`
		ScrapeFrequencyHours int        `gorm:"column:scrape_frequency_hours"`
	}

	err := app.db.Table("sites_config").
		Select("site_name, scraping_status, scraping_started_at, scraping_error, last_scraped_at, scrape_frequency_hours").
		Order("site_name").
		Find(&statuses).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	app.scrapingMu.Lock()
	defer app.scrapingMu.Unlock()

	result := make([]ScrapeStatus, 0, len(statuses))
	for _, s := range statuses {
		// Check if site is due for scraping
		isDue := false
		if s.LastScrapedAt == nil {
			isDue = true
		} else {
			dueTime := s.LastScrapedAt.Add(time.Duration(s.ScrapeFrequencyHours) * time.Hour)
			isDue = time.Now().After(dueTime)
		}

		// Check if site is currently being scraped in this process
		_, isRunning := app.scrapingProcesses[s.SiteName]

		// If database says running but not in our map, it might be stale
		finalStatus := s.ScrapingStatus
		if finalStatus == StatusRunning && !isRunning {
			finalStatus = StatusIdle
		}

		result = append(result, ScrapeStatus{
			Site:             s.SiteName,
			Status:           finalStatus,
			StartedAt:        s.ScrapingStartedAt,
			Error:            s.ScrapingError,
			LastScrapedAt:    s.LastScrapedAt,
			ScrapeFrequency:  s.ScrapeFrequencyHours,
			IsDueForScraping: isDue,
		})
	}

	c.JSON(http.StatusOK, result)
}

// resetStuckScrapingJobs resets any 'running' scraping jobs to 'failed'
// Should be called on application startup
func (app *App) resetStuckScrapingJobs() {
	err := app.db.Exec(`
		UPDATE sites_config 
		SET scraping_status = ?,
			scraping_error = 'Process terminated unexpectedly'
		WHERE scraping_status = ?
	`, StatusFailed, StatusRunning).Error

	if err != nil {
		log.Printf("Failed to reset stuck scraping jobs: %v", err)
	} else {
		log.Println("Reset stuck scraping jobs (if any)")
	}
}

// updateScrapingStatus updates scraping status for multiple sites
func (app *App) updateScrapingStatus(sites []string, status, errorMsg string) {
	for _, site := range sites {
		if err := app.db.Exec(`
			UPDATE sites_config 
			SET scraping_status = ?,
				scraping_error = ?
			WHERE site_name = ?
		`, status, errorMsg, site).Error; err != nil {
			log.Printf("Failed to update scraping status for %s: %v", site, err)
		}
	}
}

// readPipeAsync reads from a pipe asynchronously and returns a channel with the result
func readPipeAsync(reader io.Reader) <-chan string {
	ch := make(chan string, 1)
	go func() {
		data, err := io.ReadAll(reader)
		if err != nil {
			ch <- ""
		} else {
			ch <- string(data)
		}
	}()
	return ch
}
