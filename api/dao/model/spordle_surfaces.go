package model

import (
	"time"
)

const TableNameSpordleSurface = "spordle_surfaces"

type SpordleSurface struct {
	ID                  uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	VenueID             string    `gorm:"column:venue_id;not null" json:"venue_id"`
	VenueName           string    `gorm:"column:venue_name;not null" json:"venue_name"`
	VenueAddress        string    `gorm:"column:venue_address" json:"venue_address"`
	VenueCity           string    `gorm:"column:venue_city" json:"venue_city"`
	VenueRegion         string    `gorm:"column:venue_region" json:"venue_region"`
	VenueCountry        string    `gorm:"column:venue_country" json:"venue_country"`
	VenueAlias          string    `gorm:"column:venue_alias" json:"venue_alias"`
	VenueLatitude       *float64  `gorm:"column:venue_latitude" json:"venue_latitude"`
	VenueLongitude      *float64  `gorm:"column:venue_longitude" json:"venue_longitude"`
	VenuePostalCode     string    `gorm:"column:venue_postal_code" json:"venue_postal_code"`
	SurfaceID           uint      `gorm:"column:surface_id;not null" json:"surface_id"`
	SurfaceName         string    `gorm:"column:surface_name;not null" json:"surface_name"`
	SurfaceAlias        string    `gorm:"column:surface_alias" json:"surface_alias"`
	SurfaceSports       string    `gorm:"column:surface_sports" json:"surface_sports"`
	SurfaceTimeZone     string    `gorm:"column:surface_time_zone" json:"surface_time_zone"`
	SurfaceType         string    `gorm:"column:surface_type" json:"surface_type"`
	SurfaceSize         string    `gorm:"column:surface_size" json:"surface_size"`
	LivebarnSurfaceID   string    `gorm:"column:livebarn_surface_id" json:"livebarn_surface_id"`
	NumberOfGamesComing int       `gorm:"column:number_of_games_coming" json:"number_of_games_coming"`
	CreatedAt           time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (*SpordleSurface) TableName() string {
	return TableNameSpordleSurface
}
