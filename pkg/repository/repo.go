package repository

import (
	"calendar-scrapper/config"
	"calendar-scrapper/dao/model"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Repository struct {
	DB *gorm.DB
}

func NewRepository(cfg config.Config) *Repository {
	db, err := gorm.Open(mysql.Open(cfg.DbDSN))

	if err != nil {
		panic(err)
	}

	// silent log to avoid junk in csv output
	db.Logger.LogMode(logger.Silent)
	return &Repository{
		DB: db,
	}
}

// return surface from given site and location
func (r *Repository) GetMatchingSurface(site, loc string) *model.Surface {
	var siteLoc model.SitesLocation

	err := r.DB.Raw("select * from sites_locations where site=? and location=?", site, loc).Scan(&siteLoc).Error

	if err != nil {
		log.Println(err)
		return nil
	}

	if siteLoc.SurfaceID == 0 {
		return nil
	}

	var surface model.Surface
	if err = r.DB.First(&surface, siteLoc.SurfaceID).Error; err != nil {
		log.Panicln(err)
		return nil
	}
	return &surface
}

func (r *Repository) GetSitesLocation(site, loc string) (*model.SitesLocation, error) {
	var m model.SitesLocation

	err := r.DB.First(&m, "site=? and location=?", site, loc).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) GetLocation(id int) (*model.Location, error) {
	var m model.Location

	err := r.DB.First(&m, "id=?", id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) GetGthlMappings() (map[string]int, error) {
	res := []model.GthlMapping{}

	err := r.DB.Find(&res).Error
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

	err := r.DB.Find(&res).Error
	if err != nil {
		return nil, err
	}

	var m = map[string]int{}

	for _, r := range res {
		m[r.Location] = int(r.SurfaceID)
	}
	return m, nil
}

func (r *Repository) GetMhlMappings() (map[string]int, error) {
	res := []model.MhlMapping{}

	err := r.DB.Find(&res).Error
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
	if err := r.DB.First(&model.Site{}, "site=?", site).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		log.Println("inserting site", site)
		r.DB.Save(&model.Site{Site: site, URL: ""})
	}

	return &SiteRepository{
		site,
		Repository{r.DB},
	}
}

func (r *SiteRepository) ImportLocations(locations []string) error {
	var rows = []model.SitesLocation{}

	r.DB.Where("site = ?", r.site).Find(&rows)

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
		if err := r.DB.Save(&loc).Error; err != nil {
			return fmt.Errorf("save "+l+"%w", err)
		}
	}

	return nil
}

func (r *SiteRepository) ImportLoc(locations []model.SitesLocation) error {
	var rows = []model.SitesLocation{}

	r.DB.Where("site = ?", r.site).Find(&rows)

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
		if err := r.DB.Save(&l).Error; err != nil {
			return fmt.Errorf("save "+l.Location+"%w", err)
		}
	}

	switch {
	case r.site == "lugsports":
		return r.RunMatchLocationsAllStates()
	case strings.HasSuffix(r.site, "_gs"):
		return r.MatchGamesheet()

	default:
		return r.RunMatchLocations()
	}
}

func (r *SiteRepository) MatchGamesheet() error {
	var err error

	err = r.DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Exec(`UPDATE sites_locations sl, locations l SET sl.location_id=l.id WHERE
			(sl.location LIKE l.name OR locate(l.name, sl.location)>0) AND sl.site =?`, r.site).Error

		if err != nil {
			return err
		}

		query := `UPDATE sites_locations sl, locations l, surfaces s
			SET sl.surface_id=s.id
			WHERE
			sl.site=? AND sl.location_id=l.id AND s.location_id=l.id
			AND 1=(select count(*) FROM surfaces WHERE location_id=l.id)`

		err = tx.Exec(query, r.site).Error
		if err != nil {
			return err
		}

		query = `UPDATE sites_locations sl, locations l, surfaces s
			SET sl.surface_id=s.id
			WHERE
			sl.site=? AND sl.location_id=l.id AND s.location_id=l.id
			AND locate(trim(replace(sl.location, l.name, '')), s.name)>0`

		err = tx.Exec(query, r.site).Error
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (r *SiteRepository) RunMatchLocationsAllStates() error {
	var queries = []string{
		`UPDATE sites_locations SET loc= regexp_replace(location, ' (.+)','') WHERE site=?`,

		`UPDATE
			sites_locations s,
			locations l,
			provinces p
		SET
			s.location_id = l.id,
			s.match_type='postal code'
		WHERE
			l.postal_code<>'' AND
			position(l.postal_code in s.address) AND
			p.id=l.province_id AND
			s.site=? AND
			s.location_id=0`,

		`UPDATE
			sites_locations s,
			locations l,
			provinces p
		SET
			s.location_id = l.id,
			s.match_type="partial"
		WHERE
			p.id=l.province_id AND
			position(regexp_substr(address1, '^[a-zA-Z0-9]+ [a-zA-Z0-9]+') in s.address) AND
			position(left(l.postal_code,3) in s.address) AND
			site=? AND
			s.location_id=0`,

		`UPDATE
			sites_locations s,
			locations l,
			provinces p
		SET
			s.location_id = l.id,
			s.match_type='address'
		WHERE
			position(regexp_substr(address1, '^[a-zA-Z0-9]+ [a-zA-Z0-9]+') IN s.address) AND
			p.id=l.province_id AND
			s.site=? AND
			s.location_id=0`,
	}

	return r.DB.Transaction(func(db *gorm.DB) error {
		for _, q := range queries {
			if err := r.DB.Exec(q, r.site).Error; err != nil {
				return err
			}
		}
		// set surface
		db.Exec(`update sites_locations set surface=regexp_substr(location, '\\(.+\\)') where site=? AND surface=""`, r.site)
		db.Exec(`update sites_locations set surface=regexp_replace(surface, "\\(", '') where site=?`, r.site)
		db.Exec(`update sites_locations set surface=regexp_replace(surface, '\\)', '') where site=?`, r.site)
		// set surface id
		db.Exec(`update sites_locations a, locations b, surfaces s set a.surface_id=s.id where a.location_id=b.id and s.location_id=b.id and position(a.surface in REPLACE(s.name,"#", ""))<>0 and s.id is not null and a.surface<>"" and a.site=? and a.surface_id=0`, r.site)
		db.Exec(`update sites_locations s, locations l, surfaces r set s.surface_id=r.id where s.location_id=l.id and r.location_id=l.id and l.total_surfaces=1 and s.surface_id=0 and s.site=? and s.surface_id=0`, r.site)
		return nil
	})
}

func (r *SiteRepository) RunMatchLocations() error {
	var queries = []string{
		`UPDATE sites_locations SET loc= regexp_replace(location, ' (.+)','') WHERE site=?`,

		`UPDATE
			sites_locations s,
			locations l,
			provinces p
		SET
			s.location_id = l.id,
			s.match_type='postal code'
		WHERE
			l.postal_code<>'' AND
			position(l.postal_code in s.address) AND
			p.id=l.province_id AND
			p.province_name="Ontario" AND
			s.site=? AND
			s.location_id=0`,

		`UPDATE
			sites_locations s,
			locations l,
			provinces p
		SET
			s.location_id = l.id,
			s.match_type="partial"
		WHERE
			p.id=l.province_id AND
			p.province_name="Ontario" AND
			position(regexp_substr(address1, '^[a-zA-Z0-9]+ [a-zA-Z0-9]+') in s.address) AND
			position(left(l.postal_code,3) in s.address) AND
			site=? AND
			s.location_id=0`,

		`UPDATE
			sites_locations s,
			locations l,
			provinces p
		SET
			s.location_id = l.id,
			s.match_type='address'
		WHERE
			position(regexp_substr(address1, '^[a-zA-Z0-9]+ [a-zA-Z0-9]+') IN s.address) AND
			p.id=l.province_id AND
			p.province_name="Ontario" AND
			s.site=? AND
			s.location_id=0`,
	}

	return r.DB.Transaction(func(db *gorm.DB) error {
		for _, q := range queries {
			if err := r.DB.Exec(q, r.site).Error; err != nil {
				return err
			}
		}
		// set surface
		db.Exec(`update sites_locations set surface=regexp_substr(location, '\\(.+\\)') where site=? AND surface=""`, r.site)
		db.Exec(`update sites_locations set surface=regexp_replace(surface, "\\(", '') where site=?`, r.site)
		db.Exec(`update sites_locations set surface=regexp_replace(surface, '\\)', '') where site=?`, r.site)
		// set surface id
		db.Exec(`update sites_locations a, locations b, surfaces s set a.surface_id=s.id where a.location_id=b.id and s.location_id=b.id and position(a.surface in REPLACE(s.name,"#", ""))<>0 and s.id is not null and a.surface<>"" and a.site=? and a.surface_id=0`, r.site)
		db.Exec(`update sites_locations s, locations l, surfaces r set s.surface_id=r.id where s.location_id=l.id and r.location_id=l.id and l.total_surfaces=1 and s.surface_id=0 and s.site=? and s.surface_id=0`, r.site)
		return nil
	})
}

func (r *Repository) ImportEvents(site string, records []*model.Event, cutOffDate time.Time) error {
	log.Println(site, ":importing events", site, cutOffDate)
	return r.DB.Transaction(func(db *gorm.DB) error {
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

func (r *Repository) ImportMappings(site string, m map[string]int32) error {
	log.Println(site, ":importing mappings")
	var table = site + "_mappings"

	for loc, surfaceID := range m {
		err := r.DB.Exec(
			fmt.Sprintf(`INSERT INTO %s (location, surface_id) VALUES(?,?)
			ON DUPLICATE KEY UPDATE surface_id=VALUES(surface_id)`, table),
			loc, surfaceID,
		).Error

		if err != nil {
			return err
		}
	}
	return nil
}
