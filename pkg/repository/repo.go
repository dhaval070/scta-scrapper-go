package repository

import (
	"calendar-scrapper/config"
	"calendar-scrapper/dao/model"
	"fmt"
	"log"

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

// return surface from given site and location
func (r *Repository) GetMatchingSurface(site, loc string) *model.Surface {
	var siteLoc model.SitesLocation

	err := r.db.Raw("select * from sites_locations where site=? and location=?", site, loc).Scan(&siteLoc).Error

	if err != nil {
		log.Println(err)
		return nil
	}

	if siteLoc.SurfaceID == 0 {
		return nil
	}

	var surface model.Surface
	if err = r.db.First(&surface, siteLoc.SurfaceID).Error; err != nil {
		log.Panicln(err)
		return nil
	}
	return &surface
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

func (r *SiteRepository) ImportLoc(locations []model.SitesLocation) error {
	var rows = []model.SitesLocation{}

	r.db.Where("site = ?", r.site).Find(&rows)

	var m = map[string]model.SitesLocation{}
	for _, r := range rows {
		m[r.Location] = r
	}

	for _, l := range locations {
		_, ok := m[l.Location]

		if ok {
			continue
		}

		l.Site = r.site
		if err := r.db.Save(&l).Error; err != nil {
			return fmt.Errorf("save "+l.Location+"%w", err)
		}
	}

	return nil
}
