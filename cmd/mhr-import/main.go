package main

import (
	"bytes"
	"calendar-scrapper/config"
	"calendar-scrapper/pkg/repository"
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"gorm.io/gorm"
)

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

	var l = make([]any, 15)

	err = repo.DB.Transaction(func(tx *gorm.DB) error {
		for {
			line, err := r.Read()
			if err != nil {
				break
			}

			fmt.Println(line)
			if len(line) != 15 {
				log.Fatal("line len ", len(line))
			}

			if line[0] == "" || line[0] == "Add" {
				continue
			}

			for i := 0; i < 15; i += 1 {
				if line[i] == "" {
					l[i] = nil
				} else {
					l[i] = line[i]
				}
			}

			if err := repo.DB.Exec(`insert into mhr_sheet values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
				l...).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Println("failed ", err)
	}
}
