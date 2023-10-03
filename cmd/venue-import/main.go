package main

import (
	"calendar-scrapper/config"
	"calendar-scrapper/pkg/entity"
	"calendar-scrapper/pkg/repository"
	"encoding/json"
	"flag"
	"io"
	"os"

	"gorm.io/gen"
	"gorm.io/gorm"
)

func main() {
	var path string
	var err error
	flag.StringVar(&path, "path", "", "--path=<file path>")
	flag.Parse()

	if path == "" {
		panic("path is required")
	}

	config.Init("config", ".")

	cfg := config.MustReadConfig()
	repo := repository.NewRepository(cfg)
	// l := model.Location{}

	if err != nil {
		panic(err)
	}

	var js = []entity.JsonLocation{}

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
	// log.Printf("%#v\n", js[0].Surfaces)

	repo.MasterImportJson(js)
}

func genModels(db *gorm.DB) {
	g := gen.NewGenerator(gen.Config{
		OutPath: "./query",
		Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface, // generate mode
	})

	g.UseDB(db)
	g.ApplyBasic(g.GenerateAllTable()...)

	g.Execute()
}
