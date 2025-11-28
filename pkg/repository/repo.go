package repository

import (
	"calendar-scrapper/config"
	"calendar-scrapper/dao/model"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Repository struct {
	DB *gorm.DB
}

var db *gorm.DB

func NewRepository(cfg config.Config) *Repository {
	var err error
	if db == nil {
		db, err = gorm.Open(mysql.Open(cfg.DbDSN))
		if err != nil {
			panic(err)
		}
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

	err := r.DB.Raw("SELECT * FROM sites_locations WHERE site=? AND location=?", site, loc).Scan(&siteLoc).Error

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
	case strings.HasPrefix(r.site, "gs_"):
		return r.MatchGamesheet()

	default:
		return r.RunMatchLocations()
	}
}

// Matches gamesheet season locations with livebarn locations and surfaces
func (r *SiteRepository) MatchGamesheet() error {
	var err error

	err = r.DB.Transaction(func(tx *gorm.DB) error {
		var queries = []string{
			// set livebarn location name contains site location name (prefer longest match)
			`UPDATE sites_locations sl 
			JOIN (
				SELECT sl2.location, MAX(LENGTH(l2.name)) as max_len
				FROM sites_locations sl2, locations l2
				WHERE (sl2.location LIKE l2.name OR locate(l2.name, sl2.location)>0) AND sl2.site =?
				GROUP BY sl2.location
			) max_matches ON sl.location = max_matches.location
			JOIN locations l ON (sl.location LIKE l.name OR locate(l.name, sl.location)>0) AND LENGTH(l.name) = max_matches.max_len
			SET sl.location_id=l.id
			WHERE sl.site =?`,

			// set surface id if matched location has just 1 surface.
			`UPDATE sites_locations sl, locations l, surfaces s
			SET sl.surface_id=s.id
			WHERE
			sl.site=? AND sl.location_id=l.id AND s.location_id=l.id
			AND 1=(select count(*) FROM surfaces WHERE location_id=l.id)`,

			// set surface id where remaining part of surface location matches with surface name
			`UPDATE sites_locations sl, locations l, surfaces s
			SET sl.surface_id=s.id
			WHERE
			sl.site=? AND sl.location_id=l.id AND s.location_id=l.id
			AND locate(trim(replace(sl.location, l.name, '')), s.name)>0`,
		}

		// Execute first query with two site parameters
		err = tx.Exec(queries[0], r.site, r.site).Error
		if err != nil {
			return err
		}

		// Execute remaining queries with single site parameter
		for _, q := range queries[1:] {
			err = tx.Exec(q, r.site).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
	var locations []model.Location
	err = r.DB.Raw(`SELECT id, name FROM locations WHERE id NOT in(SELECT location_id FROM sites_locations WHERE site LIKE 'gs\_%')`).Scan(&locations).Error
	if err != nil {
		return err
	}
	var siteLoc []model.SitesLocation
	err = r.DB.Raw(`SELECT site, location FROM sites_locations WHERE site LIKE 'gs\_%'
		AND surface_id=0`).Scan(&siteLoc).Error
	if err != nil {
		return err
	}

	var ids = make([]int, 0, len(locations))
	for _, l := range locations {
		ids = append(ids, int(l.ID))
	}

	var surfaces = make([]model.Surface, 0, len(locations)*2)

	err = r.DB.Where("location_id in ?", ids).Find(&surfaces).Error

	smap := map[int32][]model.Surface{}

	for _, s := range surfaces {
		smap[s.LocationID] = append(smap[s.LocationID], s)
	}

	var totalLocMatch, totalSurfaceMatch = 0, 0
	for _, sl := range siteLoc {
		locMatched, surfaceMatched, err := r.MatchLocByTokens(sl, locations, smap)
		if err != nil {
			log.Println(err)
			return err
		}

		if locMatched {
			totalLocMatch++
		}
		if surfaceMatched {
			totalSurfaceMatch++
		}
	}

	log.Printf("gamesheet: by tokens totalLocMatched=%d, totalSurfaceMatched=%d\n", totalLocMatch, totalSurfaceMatch)

	return nil
}

// splits slite location words and match with each livebarn location. also sets surface id by matching last word in site location.
func (r *SiteRepository) MatchLocByTokens(sl model.SitesLocation, locations []model.Location, smap map[int32][]model.Surface) (bool, bool, error) {
	var err error
	tokens := strings.Split(sl.Location, " ")
	re := regexp.MustCompile("[^0-9A-Za-z]")

	locMatch := false
	surfaceMatch := false

	blackList := map[string]bool{
		"ice": true, "arena": true, "pavilion": true, "centennial": true,
		"arctic": true, "national": true, "sports": true, "sportplex": true,
		"bell": true, "center": true, "centre": true, "field": true,
		"fields": true, "livebarn": true, "convention": true,
	}

TOKENS_LOOP:
	for _, t := range tokens {
		if len(t) < 2 || re.MatchString(t) || blackList[strings.ToLower(t)] {
			log.Println("skipping location ", sl.Location)
			continue
		}
		var id int32 = 0

		for _, l := range locations {
			if strings.Contains(l.Name, t) {
				// if more than one location matched then skip token.
				if id != 0 {
					log.Println("multiple loc matched for site-location:", sl.Location, ",token:", t)
					continue TOKENS_LOOP
				}
				id = l.ID
			}
		}
		if id == 0 {
			continue
		}

		err = r.DB.Transaction(func(tx *gorm.DB) error {
			// set location id
			err = tx.Exec(`UPDATE sites_locations set location_id=? WHERE site=? AND location=?`, id, sl.Site, sl.Location).Error
			if err != nil {
				return fmt.Errorf("failed to set location id, %w", err)
			}
			locMatch = true

			log.Printf("gamesheet matched location. site=%s, location=%s, locId=%d, token=%s\n", sl.Site, sl.Location, id, t)

			// locaton has single surface then it is a match.
			if len(smap[id]) == 1 {
				surfaceMatch = true
				log.Printf("gamesheet matched surface: single, site=%s, location=%s\n", sl.Site, sl.Location)
				err = tx.Exec(`UPDATE sites_locations SET surface_id=? WHERE
						site=? AND location=?`, smap[id][0].ID, sl.Site, sl.Location).Error
				if err != nil {
					return fmt.Errorf("failed to set location id, %w", err)
				}
			} else {
				lastWord := tokens[len(tokens)-1]
				lastWord = re.ReplaceAllString(lastWord, "")
				if lastWord == "" {
					return nil
				}

				for _, s := range smap[id] {
					if !strings.Contains(strings.ToLower(s.Name), strings.ToLower(lastWord)) {
						continue
					}
					surfaceMatch = true
					log.Printf("gamesheet matched surface: lastword, site=%s, location=%s\n", sl.Site, sl.Location)

					err = tx.Exec(`UPDATE sites_locations sET surface_id=? WHERE
							site=? AND location=?`, s.ID, sl.Site, sl.Location).Error
					if err != nil {
						return fmt.Errorf("failed to set location id, %w", err)
					}
				}
			}
			return nil
		})
		if err != nil {
			return false, false, err
		}
		break
	}
	return locMatch, surfaceMatch, nil
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
