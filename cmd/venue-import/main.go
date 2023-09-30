package main

import (
	"calendar-scrapper/config"
	"calendar-scrapper/dao/model"
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	var path string
	flag.StringVar(&path, "path", "", "--path=<file path>")
	flag.Parse()

	if path == "" {
		panic("path is required")
	}

	config.Init("config", ".")

	cfg := config.MustReadConfig()
	l := model.Location{}

	db, err := gorm.Open(mysql.Open(cfg.DbDSN))

	if err != nil {
		panic(err)
	}

	var js = []any{}
	fh, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	data, err := io.ReadAll(fh)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(data, &js)
	if err != nil {
		panic(err)
	}
	log.Println(js[0])

	for _, loc := range js {
		importLoc(loc)
	}
}

func importLoc(l map[string]any) {
	loc := model.Location{
		ID:                  l["id"].(int32),
		Address1:            l["address1"].(string),
		City:                l["city"].(string),
		Name:                l["name"].(string),
		UUID:                l["uuid"].(string),
		RecordingHoursLocal: l["recordingHoursLocal"].(string),
		PostalCode:          l["PostalCode"].(string),
		AllSheetsCount:      l["allSheetCount"].(string),
		Longitude:           l["longitude"].(float32),
		Latitude:            l["latitude"].(float32),
		LogoURL:             l["logoUrl"].(string),
	}
}
