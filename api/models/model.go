package models

import (
	"surface-api/dao/model"
	"time"
)

type Config struct {
	DB_DSN           string `mapstructure:"DB_DSN"`
	Port             string `mapstructure:"port"`
	Mode             string `mapstructure:"mode"`
	ScraperPath      string `mapstructure:"scraper_path"`
	GameSheetAPIKey  string `mapstructure:"GAMESHEET_API_KEY"`
	GameSheetCSVPath string `mapstructure:"GAMESHEET_CSV_PATH"`
}

type SiteLoc struct {
	Site              string         `gorm:"column:site;not null" json:"site"`
	Location          string         `gorm:"column:location" json:"location"`
	LocationID        int32          `gorm:"column:location_id" json:"location_id"`
	Address           string         `gorm:"column:address" json:"address"`
	MatchType         string         `gorm:"column:match_type" json:"match_type"`
	SurfaceID         int32          `gorm:"column:surface_id;not null" json:"surface_id"`
	Surface           string         `gorm:"column:surface" json:"surface"`
	ScrapingStatus    string         `gorm:"-" json:"scraping_status,omitempty"`
	ScrapingStartedAt *time.Time     `gorm:"-" json:"scraping_started_at,omitempty"`
	ScrapingError     *string        `gorm:"-" json:"scraping_error,omitempty"`
	LiveBarnLocation  model.Location `gorm:"foreignKey:LocationID"`
	LinkedSurface     model.Surface  `gorm:"foreignKey:SurfaceID"`
}

func (*SiteLoc) TableName() string {
	return "sites_locations"
}

type MHRLocResult struct {
	Data    []MhrLocation `json:"data"`
	Page    int           `json:"page"`
	PerPage int           `json:"perPage"`
	Total   int64         `json:"total"`
}

type SiteLocResult struct {
	Data          []SiteLoc `json:"data"`
	Page          int       `json:"page"`
	PerPage       int       `json:"perPage"`
	Total         int64     `json:"total"`
	EventsMatched *int      `json:"events_matched,omitempty"`
	GamesClaimed  *int      `json:"games_claimed,omitempty"`
}

type KmasterVenueListInput struct {
	Validate          *int8  `json:"validate"`
	LivebarnVenueID   *int   `json:"livebarn_venue_id"`
	MhrVenueID        *int   `json:"mhr_venue_id"`
	VenueName         string `json:"venue_name" binding:"required"`
	Surfaces          *int   `json:"surfaces"`
	City              string `json:"city"`
	RinkAddress       string `json:"rink_address"`
	PostalCode        string `json:"postal_code"`
	ProvinceState     string `json:"province_state"`
	Country           string `json:"country"`
	CompanyNameAlt1   string `json:"company_name_alt1"`
	CompanyNameAlt2   string `json:"company_name_alt2"`
	CompanyNameAlt3   string `json:"company_name_alt3"`
	ParentCompany     string `json:"parent_company"`
	VenueType         string `json:"venue_type"`
	AccountStatus     string `json:"account_status"`
	StreamingPlatform string `json:"streaming_platform"`
	PhoneNumber       string `json:"phone_number"`
	Website           string `json:"website"`
}

type KmasterVenueListResponse struct {
	ID                     uint64 `json:"id"`
	Validate               int8   `json:"validate"`
	LivebarnVenueID        int    `json:"livebarn_venue_id"`
	MhrVenueID             int    `json:"mhr_venue_id"`
	VenueName              string `json:"venue_name"`
	Surfaces               int    `json:"surfaces"`
	City                   string `json:"city"`
	RinkAddress            string `json:"rink_address"`
	PostalCode             string `json:"postal_code"`
	ProvinceState          string `json:"province_state"`
	Country                string `json:"country"`
	CompanyNameAlt1        string `json:"company_name_alt1"`
	CompanyNameAlt2        string `json:"company_name_alt2"`
	CompanyNameAlt3        string `json:"company_name_alt3"`
	ParentCompany          string `json:"parent_company"`
	VenueType              string `json:"venue_type"`
	AccountStatus          string `json:"account_status"`
	StreamingPlatform      string `json:"streaming_platform"`
	PhoneNumber            string `json:"phone_number"`
	Website                string `json:"website"`
	CreatedAt              string `json:"created_at"`
	UpdatedAt              string `json:"updated_at"`
	LivebarnVenueIDMatched bool   `json:"livebarn_venue_id_matched"`
	MhrVenueIDMatched      bool   `json:"mhr_venue_id_matched"`
}
type KVenueResult struct {
	Data    []KmasterVenueListResponse `json:"data"`
	Page    int                        `json:"page"`
	PerPage int                        `json:"perPage"`
	Total   int64                      `json:"total"`
}

type KmasterVenueExportSurface struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type KmasterVenueExportItem struct {
	ID              uint64                      `json:"id"`
	VenueName       string                      `json:"venue_name"`
	Country         string                      `json:"country"`
	ProvinceState   string                      `json:"province_state"`
	City            string                      `json:"city"`
	LivebarnVenueID int                         `json:"livebarn_venue_id"`
	MhrVenueID      int                         `json:"mhr_venue_id"`
	Surfaces        []KmasterVenueExportSurface `json:"surfaces,omitempty"`
}

type KmasterVenueExportResult struct {
	Data  []KmasterVenueExportItem `json:"data"`
	Total int64                    `json:"total"`
}

type Mapping struct {
	Site        string `json:"site" gorm:"column:site"`
	Location    string `json:"location" gorm:"column:location"`
	SurfaceID   int    `json:"surface_id" gorm:"column:surface_id"`
	SurfaceName string `gorm:"foreignKey:SurfaceID" json:"surface_name"`
}

type SurfaceResult struct {
	ID              int32  `gorm:"column:id;primaryKey" json:"id"`
	LocationID      int32  `gorm:"column:location_id;not null" json:"location_id"`
	Name            string `gorm:"column:name;not null" json:"name"`
	Sports          string `gorm:"column:sports;not null" json:"sports"`
	LocationName    string `json:"location_name"`
	LocationCity    string `json:"location_city"`
	LocationAddress string `json:"location_address"`
}

// TableName Surface's table name
func (*SurfaceResult) TableName() string {
	return "surfaces"
}

type SetSurfaceInput struct {
	Site      string `json:"site"`
	Location  string `json:"location"`
	SurfaceID int32  `json:"surface_id"`
}

type SetLocationInput struct {
	Site       string `json:"site"`
	Location   string `json:"location"`
	LocationID int32  `json:"surface_id"`
}

type Login struct {
	Username  string     `json:"username" gorm:"primaryKey"`
	Password  string     `json:"password" gorm:"not null"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

func (Login) TableName() string {
	return "users"
}

type CreateUserInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RampLocation struct {
	Rarid        int    `json:"rar_id" gorm:"primaryKey"`
	Location     string `json:"location"`
	Address      string `json:"address"`
	City         string `json:"city"`
	ProvinceName string `json:"province_name"`
	Country      string `json:"country"`
	MatchType    string `json:"match_type"`
	SurfaceID    int    `json:"surface_id"`
	SurfaceName  string `json:"surface_name"`
}

func (RampLocation) TableName() string {
	return "RAMP_Locations"
}

type SetRampSurfaceID struct {
	RarID     int    `json:"rar_id" binding:"required"`
	SurfaceID int    `json:"surface_id"`
	Province  string `json:"province"`
}

type SitesConfigInput struct {
	SiteName             string         `json:"site_name" binding:"required"`
	DisplayName          *string        `json:"display_name"`
	BaseURL              string         `json:"base_url"`
	HomeTeam             *string        `json:"home_team"`
	ParserType           string         `json:"parser_type" binding:"required"`
	ParserConfig         map[string]any `json:"parser_config"`
	Enabled              *bool          `json:"enabled"`
	ScrapeFrequencyHours *int32         `json:"scrape_frequency_hours"`
	Notes                *string        `json:"notes"`
	ReadinessStatus      *int32         `json:"readiness_status"`
}

type SitesConfigResponse struct {
	ID                   int32          `json:"id"`
	SiteName             string         `json:"site_name"`
	DisplayName          *string        `json:"display_name"`
	BaseURL              string         `json:"base_url"`
	HomeTeam             *string        `json:"home_team"`
	ParserType           string         `json:"parser_type"`
	ParserConfig         map[string]any `json:"parser_config"`
	Enabled              *bool          `json:"enabled"`
	LastScrapedAt        *string        `json:"last_scraped_at"`
	ScrapeFrequencyHours *int32         `json:"scrape_frequency_hours"`
	Notes                *string        `json:"notes"`
	ScrapingStatus       string         `json:"scraping_status"`
	ScrapingStartedAt    *string        `json:"scraping_started_at"`
	ScrapingError        *string        `json:"scraping_error"`
	ReadinessStatus      int32          `json:"readiness_status"`
	CreatedAt            string         `json:"created_at"`
	UpdatedAt            string         `json:"updated_at"`
	GamesScraped         int32          `json:"games_scraped"`
	GamesImported        int32          `json:"games_imported"`
}

type SurfaceReport struct {
	SurfaceID    string `json:"surface_id"`
	LocationID   string `json:"location_id"`
	LocationName string `json:"location_name"`
	SurfaceName  string `json:"surface_name"`
	DayOfWeek    string `json:"day_of_week"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
}

type EventWithLocation struct {
	model.Event
	LocationName string `json:"location_name"`
	SurfaceName  string `json:"surface_name"`
	DisplayName  string `json:"display_name"`
}

type EventWithClaim struct {
	model.Event
	ClaimStatus         *int8      `json:"claim_status,omitempty"`
	ClaimHTTPStatusCode *int       `json:"claim_http_status_code,omitempty"`
	ClaimErrorMessage   *string    `json:"claim_error_message,omitempty"`
	ClaimCreatedAt      *time.Time `json:"claim_created_at,omitempty"`
	ClaimUpdatedAt      *time.Time `json:"claim_updated_at,omitempty"`
}

type EventsResult struct {
	Data    []EventWithClaim `json:"data"`
	Page    int              `json:"page"`
	PerPage int              `json:"perPage"`
	Total   int64            `json:"total"`
}

type UnsetMappingInput struct {
	Site     string `json:"site" binding:"required"`
	Location string `json:"location" binding:"required"`
	Type     string `json:"type" binding:"required,oneof=location surface"`
}

type UnsetMHRMappingInput struct {
	MhrId int    `json:"mhr_id" binding:"required"`
	Type  string `json:"type" binding:"required,oneof=location surface"`
}

type MhrLocation struct {
	MhrID              int                 `gorm:"column:mhr_id;primaryKey" json:"mhr_id"`
	RinkName           string              `gorm:"column:rink_name;not null" json:"rink_name"`
	Aka                *string             `gorm:"column:aka" json:"aka"`
	Address            string              `gorm:"column:address;not null" json:"address"`
	Phone              *string             `gorm:"column:phone" json:"phone"`
	Website            *string             `gorm:"column:website" json:"website"`
	Streaming          *string             `gorm:"column:streaming" json:"streaming"`
	Notes              *string             `gorm:"column:notes" json:"notes"`
	LivebarnInstalled  bool                `gorm:"column:livebarn_installed" json:"livebarn_installed"`
	LivebarnLocationId int                 `gorm:"column:livebarn_location_id" json:"livebarn_location_id"`
	LivebarnSurfaceId  int                 `gorm:"column:livebarn_surface_id" json:"livebarn_surface_id"`
	LiveBarnLocation   model.Location      `gorm:"foreignKey:LivebarnLocationId"`
	LinkedSurface      model.Surface       `gorm:"foreignKey:LivebarnSurfaceId"`
	HomeTeams          []map[string]string `gorm:"column:home_teams;serializer:json" json:"home_teams"`
	Province           string              `gorm:"column:province" json:"province"`
	LbNotes            string              `gorm:"column:lb_notes" json:"lb_notes"`
	CreatedAt          time.Time           `gorm:"column:created_at" json:"created_at"`
	UpdatedAt          time.Time           `gorm:"column:updated_at" json:"updated_at"`
}

func (MhrLocation) TableName() string {
	return "mhr_locations"
}
