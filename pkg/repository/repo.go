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

var blackList = map[string]bool{
	"ice": true, "arena": true, "pavilion": true, "centennial": true,
	"arctic": true, "national": true, "sports": true, "sportplex": true,
	"sportsplex": true,
	"bell":       true, "center": true, "centre": true, "field": true,
	"fields": true, "livebarn": true, "convention": true,
}

var reNonAlphaNum = regexp.MustCompile("[^0-9A-Za-z]")

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

	// Load all locations and unmatched sites_locations into memory for efficient matching
	var allLocations []model.Location
	if err = r.DB.Where("deleted_at IS NULL").Find(&allLocations).Error; err != nil {
		return err
	}

	var unmatchedSiteLocs []model.SitesLocation
	if err = r.DB.Where("site = ? AND location_id = 0", r.site).Find(&unmatchedSiteLocs).Error; err != nil {
		return err
	}

	// Match in memory - find best (longest) location name match for each site location
	type matchResult struct {
		siteLocation string
		locationID   int32
	}
	var matches []matchResult

	for _, sl := range unmatchedSiteLocs {
		var bestMatch *model.Location
		var bestLen int
		for i := range allLocations {
			loc := &allLocations[i]
			if strings.Contains(strings.ToLower(sl.Location), strings.ToLower(loc.Name)) {
				if len(loc.Name) > bestLen {
					bestMatch = loc
					bestLen = len(loc.Name)
				}
			}
		}
		if bestMatch != nil {
			matches = append(matches, matchResult{sl.Location, bestMatch.ID})
		}
	}

	log.Printf("gamesheet: in-memory matching found %d location matches\n", len(matches))

	// Batch update matched location_ids
	err = r.DB.Transaction(func(tx *gorm.DB) error {
		for _, m := range matches {
			if err := tx.Exec("UPDATE sites_locations SET location_id = ? WHERE site = ? AND location = ?",
				m.locationID, r.site, m.siteLocation).Error; err != nil {
				return err
			}
		}

		// set surface id if matched location has just 1 surface.
		err = tx.Exec(`UPDATE sites_locations sl
			JOIN surfaces s ON sl.location_id = s.location_id AND s.deleted_at IS NULL
			JOIN (SELECT location_id FROM surfaces WHERE deleted_at IS NULL GROUP BY location_id HAVING COUNT(*) = 1) single 
				ON sl.location_id = single.location_id
			SET sl.surface_id = s.id
			WHERE sl.site = ?`, r.site).Error
		if err != nil {
			return err
		}

		// set surface id where remaining part of surface location matches with surface name
		err = tx.Exec(`UPDATE sites_locations sl, locations l, surfaces s
			SET sl.surface_id=s.id
			WHERE
			sl.site=? AND sl.location_id=l.id AND s.location_id=l.id
			AND s.deleted_at IS NULL
			AND locate(s.name, trim(replace(sl.location, l.name, '')))>0`, r.site).Error
		return err
	})

	if err != nil {
		return err
	}

	var siteLoc []model.SitesLocation
	err = r.DB.Raw(`SELECT site, location, location_id FROM sites_locations WHERE site =?
		AND surface_id=0`, r.site).Scan(&siteLoc).Error
	if err != nil {
		return err
	}

	var ids = make([]int, 0, len(allLocations))
	for _, l := range allLocations {
		ids = append(ids, int(l.ID))
	}

	var surfaces = []model.Surface{}

	err = r.DB.Where("location_id in ? AND deleted_at IS NULL", ids).Find(&surfaces).Error

	smap := map[int32][]model.Surface{}

	for _, s := range surfaces {
		smap[s.LocationID] = append(smap[s.LocationID], s)
	}

	// Create location map for fast lookup
	locMap := map[int32]*model.Location{}
	for i := range allLocations {
		locMap[allLocations[i].ID] = &allLocations[i]
	}

	var totalLocMatch, totalSurfaceMatch = 0, 0
	for _, sl := range siteLoc {
		if sl.LocationID != 0 {
			id := int32(sl.LocationID)

			words := strings.Split(sl.Location, " ")
			if len(words) == 0 {
				continue
			}

			err := r.DB.Transaction(func(tx *gorm.DB) error {
				_, err = r.SetGamesheetSurface(sl, id, smap, locMap, words[len(words)-1], tx)
				return err
			})

			if err != nil {
				return err
			}
			continue
		}

		locMatched, surfaceMatched, err := r.MatchLocByTokens(sl, allLocations, smap, locMap)
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
func (r *SiteRepository) MatchLocByTokens(sl model.SitesLocation, locations []model.Location, smap map[int32][]model.Surface, locMap map[int32]*model.Location) (bool, bool, error) {
	var err error

	tokens := strings.Split(sl.Location, " ")
	lastWord := tokens[len(tokens)-1]

	locMatched := false
	surfaceMatched := false

TOKENS_LOOP:
	// match each token with livebarn location.
	for _, t := range tokens {
		if len(t) == 0 || reNonAlphaNum.MatchString(t) || blackList[strings.ToLower(t)] {
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
			locMatched = true

			log.Printf("gamesheet matched location. site=%s, location=%s, locId=%d, token=%s\n", sl.Site, sl.Location, id, t)

			surfaceMatched, err = r.SetGamesheetSurface(sl, id, smap, locMap, lastWord, tx)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return false, false, err
		}
		break
	}
	return locMatched, surfaceMatched, nil
}

func (r *SiteRepository) SetGamesheetSurface(sl model.SitesLocation, locId int32, smap map[int32][]model.Surface, locMap map[int32]*model.Location, lastWord string, tx *gorm.DB) (bool, error) {
	if len(smap[locId]) == 1 {
		return r.setSingleSurface(sl, smap[locId][0].ID, tx)
	}

	if matched, err := r.matchBySanitizedName(sl, locId, smap, locMap, tx); err != nil || matched {
		return matched, err
	}

	return r.matchByLastWord(sl, locId, smap, lastWord, tx)
}

func (r *SiteRepository) setSingleSurface(sl model.SitesLocation, surfaceID int32, tx *gorm.DB) (bool, error) {
	log.Printf("gamesheet matched surface: single, site=%s, location=%s\n", sl.Site, sl.Location)
	err := tx.Exec(`UPDATE sites_locations SET surface_id=? WHERE site=? AND location=?`,
		surfaceID, sl.Site, sl.Location).Error
	if err != nil {
		return false, fmt.Errorf("failed to set location id, %w", err)
	}
	return true, nil
}

func (r *SiteRepository) matchBySanitizedName(sl model.SitesLocation, locId int32, smap map[int32][]model.Surface, locMap map[int32]*model.Location, tx *gorm.DB) (bool, error) {
	location, ok := locMap[locId]
	if !ok {
		return false, nil
	}

	remainingPart := strings.TrimSpace(strings.ReplaceAll(sl.Location, location.Name, ""))
	sanitizedLocName := strings.ToLower(reNonAlphaNum.ReplaceAllString(remainingPart, ""))
	
	if sanitizedLocName == "" {
		return false, nil
	}

	for _, s := range smap[locId] {
		sanitizedSurfaceName := strings.ToLower(reNonAlphaNum.ReplaceAllString(s.Name, ""))
		
		if sanitizedSurfaceName != "" && strings.Contains(sanitizedLocName, sanitizedSurfaceName) {
			log.Printf("gamesheet matched surface: sanitized, site=%s, location=%s, surface=%s\n", sl.Site, sl.Location, s.Name)
			
			err := tx.Exec(`UPDATE sites_locations SET surface_id=? WHERE site=? AND location=?`,
				s.ID, sl.Site, sl.Location).Error
			if err != nil {
				return false, fmt.Errorf("failed to set surface id, %w", err)
			}
			return true, nil
		}
	}
	return false, nil
}

func (r *SiteRepository) matchByLastWord(sl model.SitesLocation, locId int32, smap map[int32][]model.Surface, lastWord string, tx *gorm.DB) (bool, error) {
	lastWord = reNonAlphaNum.ReplaceAllString(lastWord, "")
	if lastWord == "" {
		return false, nil
	}

	for _, s := range smap[locId] {
		if !strings.Contains(strings.ToLower(s.Name), strings.ToLower(lastWord)) {
			continue
		}
		
		log.Printf("gamesheet matched surface: lastword, site=%s, location=%s\n", sl.Site, sl.Location)
		err := tx.Exec(`UPDATE sites_locations SET surface_id=? WHERE site=? AND location=?`,
			s.ID, sl.Site, sl.Location).Error
		if err != nil {
			return false, fmt.Errorf("failed to set surface id, %w", err)
		}
		return true, nil
	}
	return false, nil
}

func (r *SiteRepository) RunMatchLocationsAllStates() error {
	var queries = []string{
		`UPDATE sites_locations SET loc= regexp_replace(location, ' (.+)','') WHERE site=?`,

		`UPDATE
			sites_locations s,
			locations l
		SET
			s.location_id = l.id,
			s.match_type='postal code'
		WHERE
			l.postal_code<>'' AND
			position(l.postal_code in s.address) AND
			s.site=? AND
			s.location_id=0`,

		`UPDATE
			sites_locations s,
			locations l
		SET
			s.location_id = l.id,
			s.match_type="partial"
		WHERE
			position(regexp_substr(address1, '^[a-zA-Z0-9]+ [a-zA-Z0-9]+') in s.address) AND
			position(left(l.postal_code,3) in s.address) AND
			site=? AND
			s.location_id=0`,

		`UPDATE
			sites_locations s,
			locations l
		SET
			s.location_id = l.id,
			s.match_type='address'
		WHERE
			position(regexp_substr(address1, '^[a-zA-Z0-9]+ [a-zA-Z0-9]+') IN s.address) AND
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
		db.Exec(`update sites_locations a, surfaces s set a.surface_id=s.id where s.location_id=a.location_id and position(a.surface in REPLACE(s.name,"#", ""))<>0 and s.id is not null and a.surface<>"" and a.site=? and a.surface_id=0`, r.site)

		db.Exec(`update sites_locations s, locations l, surfaces r set s.surface_id=r.id where s.location_id=l.id and r.location_id=s.location_id and l.total_surfaces=1 and s.surface_id=0 and s.site=?`, r.site)
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
		db.Exec(`update sites_locations a, surfaces s set a.surface_id=s.id where s.location_id=a.location_id and position(a.surface in REPLACE(s.name,"#", ""))<>0 and s.id is not null and a.surface<>"" and a.site=? and a.surface_id=0`, r.site)
		db.Exec(`update sites_locations s, locations l, surfaces r set s.surface_id=r.id where s.location_id=l.id and r.location_id=s.location_id and l.total_surfaces=1 and s.surface_id=0 and s.site=?`, r.site)
		return nil
	})
}

func (r *Repository) ImportEvents(site string, records []*model.Event, cutOffDate time.Time) error {
	log.Println(site, ":importing events", site, cutOffDate)

	// retry to avoid dead lock by multiple sites events import in parallel. that is done by run.sh line: ./bin/site-schedule -site $site -infile $f > $d1/$site.csv --import -cutoffdate $dt &
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := r.DB.Transaction(func(db *gorm.DB) error {
			if err := db.Exec("DELETE FROM events WHERE site=? AND datetime > ?", site, cutOffDate).Error; err != nil {
				return err
			}
			return db.CreateInBatches(records, 50).Error
		})

		if err == nil {
			return nil
		}

		// Check if it's a deadlock error (MySQL error 1213)
		if strings.Contains(err.Error(), "Deadlock") || strings.Contains(err.Error(), "1213") {
			log.Printf("%s: deadlock detected, attempt %d/%d, retrying...\n", site, attempt, maxRetries)
			time.Sleep(time.Duration(attempt*100) * time.Millisecond)
			continue
		}
		return err
	}
	return fmt.Errorf("%s: failed after %d retries due to deadlock", site, maxRetries)
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
