package main

import (
	"encoding/csv"
	"flag"
	"log"
	"os"
	"strconv"
	"strings"

	"calendar-scrapper/config"
	"calendar-scrapper/pkg/repository"

	"gorm.io/gorm"
)

type KmasterVenue struct {
	ID        uint64   `gorm:"column:id"`
	Latitude  *float64 `gorm:"column:latitude"`
	Longitude *float64 `gorm:"column:longitude"`
}

func (KmasterVenue) TableName() string {
	return "kmaster_venue_list"
}

func main() {
	var path string
	var dryRun bool
	flag.StringVar(&path, "path", "", "--path=<csv file path>")
	flag.BoolVar(&dryRun, "dry-run", false, "--dry-run (preview only, no writes)")
	flag.Parse()

	if path == "" {
		log.Fatal("path is required")
	}

	config.Init("config", ".")
	cfg := config.MustReadConfig()
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

	var updated, skipped int
	err = repo.DB.Transaction(func(tx *gorm.DB) error {
		for {
			line, err := r.Read()
			if err != nil {
				break
			}
			if len(line) < 7 {
				log.Printf("row %s: too few columns (%d), skipping", line[0], len(line))
				skipped++
				continue
			}

			idStr := strings.TrimSpace(line[1])
			latStr := strings.TrimSpace(line[5])
			lngStr := strings.TrimSpace(line[6])

			id, err := strconv.ParseUint(idStr, 10, 64)
			if err != nil {
				log.Printf("row %s: invalid id %q, skipping", line[0], idStr)
				skipped++
				continue
			}

			if latStr == "" || lngStr == "" {
				log.Printf("id %d: empty lat/lng, skipping", id)
				skipped++
				continue
			}

			lat, err := strconv.ParseFloat(latStr, 64)
			if err != nil {
				log.Printf("id %d: invalid latitude %q, skipping", id, latStr)
				skipped++
				continue
			}

			lng, err := strconv.ParseFloat(lngStr, 64)
			if err != nil {
				log.Printf("id %d: invalid longitude %q, skipping", id, lngStr)
				skipped++
				continue
			}

			if dryRun {
				log.Printf("[dry-run] would update id %d: lat=%f, lng=%f", id, lat, lng)
				updated++
				continue
			}

			if err := tx.Model(&KmasterVenue{}).Where("id = ?", id).
				Updates(map[string]any{"latitude": lat, "longitude": lng}).Error; err != nil {
				return err
			}
			updated++
		}
		return nil
	})

	if err != nil {
		log.Fatalf("update failed: %v", err)
	}

	mode := ""
	if dryRun {
		mode = " (dry-run)"
	}
	log.Printf("updated %d rows, skipped %d%s", updated, skipped, mode)
}
