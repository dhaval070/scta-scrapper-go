package main

import (
	"calendar-scrapper/config"
	"log"
	"time"

	"calendar-scrapper/dao/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	config.Init("config", ".")

	cfg := config.MustReadConfig()
	db, err := gorm.Open(mysql.Open(cfg.DbDSN))

	if err != nil {
		panic(err)
	}

	db.AutoMigrate(
		&model.Event{},
		&model.Site{},
		&model.Zone{},
		&model.Surface{},
		&model.FeedMode{},
		&model.Location{},
		&model.Province{},
		&model.Rendition{},
		&model.VenueStatus{},
		&model.SitesLocation{},
		&model.SurfaceFeedMode{},
	)
	d, err := db.DB()
	if err != nil {
		panic(err)
	}

	loc, err := time.LoadLocation("UTC")
	if err != nil {
		panic(err)
	}

	log.Println("loc", loc)

	res, err := d.Query("select now()")
	if err != nil {
		panic(err)
	}

	var dt time.Time
	log.Println(err)
	if res.Next() {
		s := res.Scan(&dt)
		log.Println(s, dt)
	}
}
