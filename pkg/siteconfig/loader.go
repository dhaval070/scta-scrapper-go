package siteconfig

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Loader handles loading site configurations from the database
type Loader struct {
	db *gorm.DB
}

// NewLoader creates a new configuration loader
func NewLoader(db *gorm.DB) *Loader {
	return &Loader{db: db}
}

// GetSite retrieves a specific site configuration by name
func (l *Loader) GetSite(siteName string) (*SiteConfig, error) {
	var config SiteConfig
	err := l.db.Where("site_name = ? AND enabled = ?", siteName, true).First(&config).Error
	if err != nil {
		return nil, fmt.Errorf("failed to load site config for '%s': %w", siteName, err)
	}
	return &config, nil
}

// GetAllEnabled retrieves all enabled site configurations
func (l *Loader) GetAllEnabled() ([]SiteConfig, error) {
	var configs []SiteConfig
	err := l.db.Where("enabled = ?", true).Order("site_name").Find(&configs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to load enabled sites: %w", err)
	}
	return configs, nil
}

// GetByParserType retrieves all enabled sites using a specific parser type
func (l *Loader) GetByParserType(parserType string) ([]SiteConfig, error) {
	var configs []SiteConfig
	err := l.db.Where("parser_type = ? AND enabled = ?", parserType, true).
		Order("site_name").
		Find(&configs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to load sites by parser type '%s': %w", parserType, err)
	}
	return configs, nil
}

// GetParserConfig unmarshals the JSON parser configuration
func (l *Loader) GetParserConfig(site *SiteConfig) (*ParserConfigJSON, error) {
	if site.ParserConfig == "" {
		return &ParserConfigJSON{}, nil
	}
	
	var cfg ParserConfigJSON
	err := json.Unmarshal([]byte(site.ParserConfig), &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config for site '%s': %w", site.SiteName, err)
	}
	return &cfg, nil
}

// UpdateLastScraped updates the last_scraped_at timestamp for a site
func (l *Loader) UpdateLastScraped(siteID int) error {
	now := time.Now()
	err := l.db.Model(&SiteConfig{}).
		Where("id = ?", siteID).
		Update("last_scraped_at", now).Error
	if err != nil {
		return fmt.Errorf("failed to update last_scraped_at for site %d: %w", siteID, err)
	}
	return nil
}

// GetDueForScraping returns sites that haven't been scraped within their frequency window
func (l *Loader) GetDueForScraping() ([]SiteConfig, error) {
	var configs []SiteConfig
	
	// Sites that have never been scraped OR last scraped more than frequency hours ago
	err := l.db.Where("enabled = ?", true).
		Where("last_scraped_at IS NULL OR last_scraped_at < DATE_SUB(NOW(), INTERVAL scrape_frequency_hours HOUR)").
		Order("last_scraped_at ASC NULLS FIRST").
		Find(&configs).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to load sites due for scraping: %w", err)
	}
	return configs, nil
}

// DisableSite disables a site (sets enabled = false)
func (l *Loader) DisableSite(siteName string) error {
	err := l.db.Model(&SiteConfig{}).
		Where("site_name = ?", siteName).
		Update("enabled", false).Error
	if err != nil {
		return fmt.Errorf("failed to disable site '%s': %w", siteName, err)
	}
	return nil
}

// EnableSite enables a site (sets enabled = true)
func (l *Loader) EnableSite(siteName string) error {
	err := l.db.Model(&SiteConfig{}).
		Where("site_name = ?", siteName).
		Update("enabled", true).Error
	if err != nil {
		return fmt.Errorf("failed to enable site '%s': %w", siteName, err)
	}
	return nil
}
