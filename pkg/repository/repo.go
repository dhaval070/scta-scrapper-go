package repository

import (
	"calendar-scrapper/config"
	"calendar-scrapper/dao/model"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(cfg config.Config) *Repository {
	db, err := gorm.Open(mysql.Open(cfg.DbDSN))

	if err != nil {
		panic(err)
	}

	return &Repository{
		db: db,
	}
}

type SiteRepository struct {
	site string
	Repository
}

func (r *Repository) Site(site string) *SiteRepository {
	return &SiteRepository{
		site,
		Repository{r.db},
	}
}

func (r *SiteRepository) ImportLocations(locations []string) error {
	var rows = []model.SitesLocation{}

	r.db.Where("site = ?", r.site).Find(&rows)

	var m = map[string]model.SitesLocation{}
	for _, r := range rows {
		m[r.Location] = r
	}

	for _, l := range locations {
		loc, ok := m[l]

		if ok {
			continue
		}

		loc = model.SitesLocation{
			Site:       r.site,
			Location:   l,
			LocationID: 0,
		}
		if err := r.db.Save(&loc).Error; err != nil {
			return fmt.Errorf("save "+l+"%w", err)
		}
	}

	return nil
}
