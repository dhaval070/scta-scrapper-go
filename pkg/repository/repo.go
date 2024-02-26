package repository

import (
	"calendar-scrapper/config"
	"calendar-scrapper/dao/model"
	"fmt"
	"log"
	"time"

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

func (r *Repository) GetGthlMappings() (map[string]int, error) {
	res := []model.GthlMapping{}

	err := r.db.Find(&res).Error
	if err != nil {
		return nil, err
	}

	var m = map[string]int{}

	for _, r := range res {
		m[r.Location] = int(r.SurfaceID)
	}
	return m, nil
}

func (r *Repository) GetNyhlMappings() (map[string]int, error) {
	res := []model.NyhlMapping{}

	err := r.db.Find(&res).Error
	if err != nil {
		return nil, err
	}

	var m = map[string]int{}

	for _, r := range res {
		m[r.Location] = int(r.SurfaceID)
	}
	return m, nil
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

		"update sites_locations s, locations l set s.location_id = l.id, s.match_type='postal code' where l.postal_code<>'' and position(l.postal_code in s.address) and s.site=? and s.location_id=0",

		`update sites_locations s, locations l set s.location_id = l.id, s.match_type="partial" where position(regexp_substr(address1, '^[a-zA-Z0-9]+ [a-zA-Z0-9]+') in s.address) and position(left(l.postal_code,3) in s.address) and site=? and s.location_id=0`,

		"update sites_locations s, locations l set s.location_id = l.id, s.match_type='address' where position(regexp_substr(address1, '^[a-zA-Z0-9]+ [a-zA-Z0-9]+') in s.address) and s.site=? and s.location_id=0",
	}

	return r.db.Transaction(func(db *gorm.DB) error {
		for _, q := range queries {
			if err := r.db.Exec(q, r.site).Error; err != nil {
				return err
			}
		}
		// set surface
		db.Exec(`update sites_locations set surface=regexp_substr(location, '\\(.+\\)') where site=?`, r.site)
		db.Exec(`update sites_locations set surface=regexp_replace(surface, "\\(", '') where site=?`, r.site)
		db.Exec(`update sites_locations set surface=regexp_replace(surface, '\\)', '') where site=?`, r.site)
		// set surface id
		db.Exec(`update sites_locations a, locations b, surfaces s set a.surface_id=s.id where a.location_id=b.id and s.location_id=b.id and position(a.surface in s.name)<>0 and s.id is not null and a.surface<>"" and a.site=? and a.surface_id=0`, r.site)
		db.Exec(`update sites_locations s, locations l, surfaces r set s.surface_id=r.id where s.location_id=l.id and r.location_id=l.id and l.total_surfaces=1 and s.surface_id=0 and s.site=? and s.surface_id=0`, r.site)
		return nil
	})
}

func (r *Repository) ImportEvents(site string, records []*model.Event, cutOffDate time.Time) error {
	log.Println(site, cutOffDate)
	return r.db.Transaction(func(db *gorm.DB) error {
		if err := db.Exec("delete from events where site=? and datetime > ?", site, cutOffDate).Error; err != nil {
			return err
		}
		var err error
		for _, rec := range records {
			if err = db.Create(rec).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
