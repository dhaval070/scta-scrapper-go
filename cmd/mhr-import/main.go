package main

import (
	"bytes"
	"calendar-scrapper/config"
	"calendar-scrapper/pkg/repository"
	"encoding/csv"
	"log"
	"os"
	"strconv"

	"gorm.io/gorm"
)

type Venue struct {
	ID                int    `gorm:"column:id;primaryKey"`
	RinkName          string `gorm:"column:rink_name;not null"`
	LivebarnInstalled bool   `gorm:"column:livebarn_installed"`
	LivebarnVenueID   int    `gorm:"column:livebarn_venue_id"`
	RinkType          string `gorm:"column:rink_type"`
	AltName           string `gorm:"column:alt_name"`
	AltName2          string `gorm:"column:alt_name2"`
	AltName3          string `gorm:"column:alt_name3"`
	RinkPads          int    `gorm:"column:rink_pads"`
	City              string `gorm:"column:city;not null"`
	State             string `gorm:"column:state;not null"`
	Address           string `gorm:"column:address"`
	Zip               string `gorm:"column:zip"`
	Country           string `gorm:"column:country"`
	Phone             string `gorm:"column:phone"`
}

func (Venue) TableName() string {
	return "mhr_sheet"
}

func main() {
	config.Init("config", ".")

	var cfg = config.MustReadConfig()
	repo := repository.NewRepository(cfg)

	path := "mhr-rinks.csv"
	b, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	br := bytes.NewReader(b)
	r := csv.NewReader(br)
	r.Read()

	err = repo.DB.Transaction(func(tx *gorm.DB) error {
		for {
			line, err := r.Read()
			if err != nil {
				break
			}

			// fmt.Println(line)
			if len(line) != 15 {
				log.Fatal("line len ", len(line))
			}

			if line[0] == "" || line[0] == "Add" {
				continue
			}

			id, err := strconv.Atoi(line[0])
			if err != nil {
				log.Println("failed to convert id ", line[0], err)
				continue
			}
			lbInstalled := false

			if line[2] == "1" {
				lbInstalled = true
			}

			var lbId int
			if line[3] != "" {
				lbId, err = strconv.Atoi(line[3])
				if err != nil {
					log.Println("failed to convert livebarn id ", line[3], err)
					continue
				}
			}

			var rinkPads int
			if line[8] != "" {
				rinkPads, err = strconv.Atoi(line[8])
				if err != nil {
					log.Println("failed to convert rinkpads ", line[8], err)
					continue
				}
			}

			row := Venue{
				ID:                id,
				RinkName:          line[1],
				LivebarnInstalled: lbInstalled,
				LivebarnVenueID:   lbId,
				RinkType:          line[4],
				AltName:           line[5],
				AltName2:          line[6],
				AltName3:          line[7],
				RinkPads:          rinkPads,
				City:              line[9],
				State:             line[10],
				Address:           line[11],
				Zip:               line[12],
				Country:           line[13],
				Phone:             line[14],
			}
			if err = tx.Save(row).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Println("failed ", err)
	}
}
