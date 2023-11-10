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

func (r *Repository) GetSitesLocation(site, loc string) (*model.SitesLocation, error) {
	var m model.SitesLocation

	err := r.db.First(&m, "site=? and location=?", site, loc).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) GetLocation(id int) (*model.Location, error) {
	var m model.Location

	err := r.db.First(&m, "id=?", id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
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

	return r.RunMatchLocations()
}

func (r *SiteRepository) RunMatchLocations() error {
	var queries = []string{
		"update sites_locations set loc= regexp_replace(location, ' (.+)','') where site=?",
		"update sites_locations set location_id = null, match_type = null where site=?",
		"update sites_locations s, locations l set s.location_id = l.id, s.match_type='postal code' where l.postal_code<>'' and position(l.postal_code in s.address) and s.site=?",
		"update sites_locations s, locations l set s.location_id = l.id, s.match_type='address' where position(regexp_substr(address1, '^[a-zA-Z0-9]+ [a-zA-Z0-9]+') in s.address) and s.location_id is null and s.site=?",
		"update sites_locations a,locations b set a.location_id=b.id,match_type='name' where position(a.loc in b.name) and a.location_id is null and a.site=?",
	}

	return r.db.Transaction(func(db *gorm.DB) error {
		for _, q := range queries {
			if err := r.db.Exec(q, r.site).Error; err != nil {
				return err
			}
		}
		db.Exec(`update sites_locations set surface=regexp_substr(location, '\\(.+\\)') where  site=?`, r.site)
		db.Exec(`update sites_locations set surface=regexp_replace(surface, "\\(", '') where site=?`, r.site)
		db.Exec(`update sites_locations set surface=regexp_replace(surface, '\\)', '') where site=?`, r.site)
		db.Exec(`update sites_locations a, locations b, surfaces s set a.surface_id=s.id where a.location_id=b.id and a.surface<>"" and position(a.surface in s.name)<>0 and s.id is not null and a.site=?`, r.site)
		return nil
	})
}
