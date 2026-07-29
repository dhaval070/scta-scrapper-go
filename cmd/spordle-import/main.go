package main

import (
	"calendar-scrapper/config"
	"calendar-scrapper/pkg/repository"
	"encoding/csv"
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type SpordleSurface struct {
	ID                  uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	VenueID             string    `gorm:"column:venue_id;not null"`
	VenueName           string    `gorm:"column:venue_name;not null"`
	VenueAddress        string    `gorm:"column:venue_address"`
	VenueCity           string    `gorm:"column:venue_city"`
	VenueRegion         string    `gorm:"column:venue_region"`
	VenueCountry        string    `gorm:"column:venue_country"`
	VenueAlias          string    `gorm:"column:venue_alias"`
	VenueLatitude       float64   `gorm:"column:venue_latitude"`
	VenueLongitude      float64   `gorm:"column:venue_longitude"`
	VenuePostalCode     string    `gorm:"column:venue_postal_code"`
	SurfaceID           uint      `gorm:"column:surface_id;not null"`
	SurfaceName         string    `gorm:"column:surface_name;not null"`
	SurfaceAlias        string    `gorm:"column:surface_alias"`
	SurfaceSports       string    `gorm:"column:surface_sports"`
	SurfaceTimeZone     string    `gorm:"column:surface_time_zone"`
	SurfaceType         string    `gorm:"column:surface_type"`
	SurfaceSize         string    `gorm:"column:surface_size"`
	LivebarnSurfaceID   string    `gorm:"column:livebarn_surface_id"`
	NumberOfGamesComing int       `gorm:"column:number_of_games_coming"`
	CreatedAt           time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt           time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP"`
}

func (SpordleSurface) TableName() string {
	return "spordle_surfaces"
}

func main() {
	var path string
	flag.StringVar(&path, "path", "", "--path=<csv file path>")
	flag.Parse()

	if path == "" {
		log.Fatal("path is required")
	}

	config.Init("config", ".")

	var cfg = config.MustReadConfig()
	repo := repository.NewRepository(cfg)

	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer f.Close()
	r := csv.NewReader(f)

	if _, err := r.Read(); err != nil {
		log.Fatalf("failed to read header: %v", err)
	}

	var count int
	err = repo.DB.Transaction(func(tx *gorm.DB) error {
		for {
			line, err := r.Read()
			if err != nil {
				break
			}

			if len(line) < 19 {
				log.Printf("skipping row with %d columns", len(line))
				continue
			}

			record := SpordleSurface{
				VenueID:         line[0],
				VenueName:       line[1],
				VenueAddress:    line[2],
				VenueCity:       line[3],
				VenueRegion:     line[4],
				VenueCountry:    line[5],
				VenueAlias:      line[6],
				SurfaceID:       parseUint(line[10]),
				SurfaceName:     line[11],
				SurfaceAlias:    line[12],
				SurfaceSports:   line[13],
				SurfaceTimeZone: line[14],
				SurfaceType:     line[15],
				SurfaceSize:     line[16],
				LivebarnSurfaceID: line[17],
			}

			if line[7] != "" {
				v, err := strconv.ParseFloat(line[7], 64)
				if err == nil {
					record.VenueLatitude = v
				}
			}
			if line[8] != "" {
				v, err := strconv.ParseFloat(line[8], 64)
				if err == nil {
					record.VenueLongitude = v
				}
			}
			record.VenuePostalCode = line[9]
			if line[18] != "" {
				v, err := strconv.Atoi(line[18])
				if err == nil {
					record.NumberOfGamesComing = v
				}
			}

			if err := tx.Save(&record).Error; err != nil {
				return err
			}
			count++
		}
		return nil
	})

	if err != nil {
		log.Fatalf("import failed: %v", err)
	}

	log.Printf("imported %d records", count)
}

func parseUint(s string) uint {
	if s == "" {
		return 0
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return uint(v)
}
