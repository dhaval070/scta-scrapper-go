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

type KmasterVenue struct {
	ID                uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	Validate          int8      `gorm:"column:validate;not null;default:0"`
	LivebarnVenueID   int       `gorm:"column:livebarn_venue_id"`
	MhrVenueID        int       `gorm:"column:mhr_venue_id"`
	VenueName         string    `gorm:"column:venue_name;not null"`
	Surfaces          int       `gorm:"column:surfaces;not null;default:0"`
	City              string    `gorm:"column:city"`
	RinkAddress       string    `gorm:"column:rink_address"`
	PostalCode        string    `gorm:"column:postal_code"`
	ProvinceState     string    `gorm:"column:province_state"`
	Country           string    `gorm:"column:country"`
	CompanyNameAlt1   string    `gorm:"column:company_name_alt1"`
	CompanyNameAlt2   string    `gorm:"column:company_name_alt2"`
	CompanyNameAlt3   string    `gorm:"column:company_name_alt3"`
	ParentCompany     string    `gorm:"column:parent_company"`
	VenueType         string    `gorm:"column:venue_type"`
	AccountStatus     string    `gorm:"column:account_status"`
	StreamingPlatform string    `gorm:"column:streaming_platform"`
	PhoneNumber       string    `gorm:"column:phone_number"`
	Website           string    `gorm:"column:website"`
	CreatedAt         time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt         time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP"`
}

func (KmasterVenue) TableName() string {
	return "kmaster_venue_list"
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

	// Read and skip header row
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

			// CSV columns: Validate,LiveBarn Venue ID,MHR Venue #,Venue Name,Surfaces,City,rink_address,Postal Code,Province State (abrv),Country (abrv),Company Name Alt 1,Company Name Alt 2,Company Name Alt 3,Parent Company,Venue Type,Account Status,Streaming Platform,Phone Number,Website
			if len(line) < 19 {
				log.Printf("skipping row with %d columns", len(line))
				continue
			}

			record := KmasterVenue{
				VenueName:         line[3],
				City:              line[5],
				RinkAddress:       line[6],
				PostalCode:        line[7],
				ProvinceState:     line[8],
				Country:           line[9],
				CompanyNameAlt1:   line[10],
				CompanyNameAlt2:   line[11],
				CompanyNameAlt3:   line[12],
				ParentCompany:     line[13],
				VenueType:         line[14],
				AccountStatus:     line[15],
				StreamingPlatform: line[16],
				PhoneNumber:       line[17],
				Website:           line[18],
			}

			if line[0] != "" {
				v, err := strconv.Atoi(line[0])
				if err == nil {
					record.Validate = int8(v)
				}
			}

			if line[1] != "" {
				id, err := strconv.Atoi(line[1])
				if err == nil {
					record.LivebarnVenueID = id
				}
			}

			if line[2] != "" {
				id, err := strconv.Atoi(line[2])
				if err == nil {
					record.MhrVenueID = id
				}
			}

			if line[4] != "" {
				s, err := strconv.Atoi(line[4])
				if err == nil {
					record.Surfaces = s
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
