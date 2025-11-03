package siteconfig

import "time"

// SiteConfig represents the configuration for a scraper site
type SiteConfig struct {
	ID                   int        `gorm:"primaryKey"`
	SiteName             string     `gorm:"uniqueIndex;not null"`
	DisplayName          string     `gorm:"type:varchar(200)"`
	BaseURL              string     `gorm:"type:varchar(500);not null"`
	HomeTeam             string     `gorm:"type:varchar(100)"`
	ParserType           string     `gorm:"type:enum('day_details','day_details_parser1','day_details_parser2','month_based','group_based','custom','external');not null"`
	ParserConfig         string     `gorm:"type:json"`
	Enabled              bool       `gorm:"default:true"`
	LastScrapedAt        *time.Time `gorm:"type:timestamp null"`
	ScrapeFrequencyHours int        `gorm:"default:24"`
	Notes                string     `gorm:"type:text"`
	CreatedAt            time.Time  `gorm:"autoCreateTime"`
	UpdatedAt            time.Time  `gorm:"autoUpdateTime"`
}

// TableName specifies the table name for GORM
func (SiteConfig) TableName() string {
	return "sites_config"
}

// ParserConfigJSON represents the JSON configuration for parsers
type ParserConfigJSON struct {
	// Common fields
	TournamentCheckExact bool   `json:"tournament_check_exact,omitempty"`
	LogErrors            bool   `json:"log_errors,omitempty"`
	URLTemplate          string `json:"url_template,omitempty"`
	ContentFilter        string `json:"content_filter,omitempty"` // Filter events by content substring
	
	// Group-based parser fields
	GroupXPath        string `json:"group_xpath,omitempty"`
	GroupURLTemplate  string `json:"group_url_template,omitempty"`
	SeasonsURL        string `json:"seasons_url,omitempty"`
	
	// Month-based parser fields
	TeamParseStrategy string `json:"team_parse_strategy,omitempty"`
	URLPrefix         string `json:"url_prefix,omitempty"`
	
	// External parser fields
	BinaryPath string   `json:"binary_path,omitempty"`
	ExtraArgs  []string `json:"extra_args,omitempty"`
}

// Parser type constants
const (
	ParserTypeDayDetails        = "day_details"
	ParserTypeDayDetailsParser1 = "day_details_parser1"
	ParserTypeDayDetailsParser2 = "day_details_parser2"
	ParserTypeMonthBased        = "month_based"
	ParserTypeGroupBased        = "group_based"
	ParserTypeCustom            = "custom"
	ParserTypeExternal          = "external"
)
