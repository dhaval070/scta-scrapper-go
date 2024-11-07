package main

import (
	"calendar-scrapper/config"
	"calendar-scrapper/pkg/repository"
	"encoding/json"
	"log"
	"os"
)

func main() {
	config.Init("config", ".")
	cfg := config.MustReadConfig()
	repo := repository.NewRepository(cfg)

	b, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	var data map[string]string

	err = json.Unmarshal(b, &data)
	if err != nil {
		panic(err)
	}

	for k, v := range data {
		log.Println(k, v)
		if err = repo.DB.Exec(`update RAMP_Locations set province_name=? where prov=?`, v, k).Error; err != nil {
			panic(err)
		}
	}
}
