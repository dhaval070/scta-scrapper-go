package main

import (
	"calendar-scrapper/config"
	"calendar-scrapper/pkg/gamesheet"
	"flag"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	flag.Parse()

	var err error

	config.Init("config", ".")
	cfg := config.MustReadConfig()

	_, err = initDB(&cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	seasons, err := gamesheet.FetchActiveSeasons(cfg.GameSheetAPIKey)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("Active seasons: %+v\n", seasons)
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DbDSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
