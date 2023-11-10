package locations

import (
	"calendar-scrapper/config"
	"calendar-scrapper/dao/model"
	"calendar-scrapper/pkg/repository"
	"fmt"
	"log"
)

func Import(site string, result [][]string) {
	var err error
	config.Init("config", ".")
	cfg := config.MustReadConfig()

	var locations = make([]model.SitesLocation, 0, len(result))
	for _, r := range result {
		log.Printf("%+v\n", r)

		l := model.SitesLocation{
			Location: r[4],
			Address:  r[6],
		}
		locations = append(locations, l)
	}

	repo := repository.NewRepository(cfg).Site(site)
	if err = repo.ImportLoc(locations); err != nil {
		log.Fatal(err)
	}
	fmt.Println("vim-go")
}
