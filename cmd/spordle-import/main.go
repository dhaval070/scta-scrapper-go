package main

import (
	"calendar-scrapper/config"
	"calendar-scrapper/pkg/repository"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"gorm.io/gorm"
)

type SpordleSurface struct {
	ID                  uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	VenueID             string    `gorm:"column:venue_id;not null"`
	VenueName           string    `gorm:"column:venue_name;not null"`
	VenueAddress        string    `gorm:"column:venue_address"`
	VenueCity           string    `gorm:"column:venue_city"`
	VenueRegion         string    `gorm:"column:venue_region"`
	VenueCountry        string    `gorm:"column:venue_country"`
	VenueAlias          string    `gorm:"column:venue_alias"`
	VenueLatitude       float64   `gorm:"column:venue_latitude"`
	VenueLongitude      float64   `gorm:"column:venue_longitude"`
	VenuePostalCode     string    `gorm:"column:venue_postal_code"`
	SurfaceID           uint      `gorm:"column:surface_id;not null"`
	SurfaceName         string    `gorm:"column:surface_name;not null"`
	SurfaceAlias        string    `gorm:"column:surface_alias"`
	SurfaceSports       string    `gorm:"column:surface_sports"`
	SurfaceTimeZone     string    `gorm:"column:surface_time_zone"`
	SurfaceType         string    `gorm:"column:surface_type"`
	SurfaceSize         string    `gorm:"column:surface_size"`
	LivebarnSurfaceID   string    `gorm:"column:livebarn_surface_id"`
	NumberOfGamesComing int       `gorm:"column:number_of_games_coming"`
	CreatedAt           time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt           time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP"`
}

func (SpordleSurface) TableName() string {
	return "spordle_surfaces"
}

type KmasterVenue struct {
	ID             uint64 `gorm:"column:id"`
	VenueName      string `gorm:"column:venue_name"`
	RinkAddress    string `gorm:"column:rink_address"`
	City           string `gorm:"column:city"`
	ProvinceState  string `gorm:"column:province_state"`
	SpordleVenueID string `gorm:"column:spordle_venue_id"`
}

func (KmasterVenue) TableName() string {
	return "kmaster_venue_list"
}

type spordleVenue struct {
	VenueID string
	Name    string
	Address string
	City    string
	Region  string
}

type kmasterUpdate struct {
	ID      uint64
	VenueID string
}

type reportRow struct {
	ID            uint64
	VenueName     string
	RinkAddress   string
	City          string
	ProvinceState string
	Reason        string
}

// provinceFixes corrects non two-letter spordle venue_region values so that
// strict province equality can be used during matching. Idempotent.
var provinceFixes = []string{
	`UPDATE spordle_surfaces SET venue_region='ON' WHERE venue_region='Ontario'`,
	`UPDATE spordle_surfaces SET venue_region='QC' WHERE venue_region='QUÉBEC'`,
	`UPDATE spordle_surfaces SET venue_region='NB' WHERE venue_region='New Brunwick'`,
	`UPDATE spordle_surfaces SET venue_region='YT' WHERE venue_region='YK'`,
	`UPDATE spordle_surfaces SET venue_region='QC' WHERE venue_region IN ('J5K 2E5','J9H 7S6')`,
}

var streetSuffixes = map[string]string{
	"rd": "road", "rte": "route", "rue": "street", "ave": "avenue",
	"blvd": "boulevard", "dr": "drive", "hwy": "highway", "ln": "lane",
	"ct": "court", "pkwy": "parkway", "pl": "place", "ter": "terrace",
}

var cityExpand = map[string]string{
	"st": "saint", "ste": "sainte",
}

var addressStopwords = map[string]bool{
	"de": true, "du": true, "des": true, "la": true, "le": true, "les": true,
	"the": true, "of": true, "sur": true, "au": true, "aux": true, "et": true,
	"and": true,
}

var cityStopwords = map[string]bool{
	"de": true, "du": true, "des": true, "la": true, "le": true, "les": true,
	"the": true, "of": true, "sur": true, "au": true, "aux": true, "et": true,
}

const (
	cityThreshold    = 0.8
	addressThreshold = 0.6
)

func main() {
	var path string
	var matchOnly, force, dryRun bool
	var report string
	flag.StringVar(&path, "path", "", "--path=<csv file path>")
	flag.BoolVar(&matchOnly, "match", false, "--match (match kmaster venues to spordle venues; standalone if no --path)")
	flag.BoolVar(&force, "force", false, "--force (re-match and overwrite existing spordle_venue_id)")
	flag.BoolVar(&dryRun, "dry-run", false, "--dry-run (preview matches only, no writes)")
	flag.StringVar(&report, "report", "", "--report=<csv file path> (write unmatched/ambiguous/skipped kmaster records)")
	flag.Parse()

	if path == "" && !matchOnly {
		log.Fatal("either --path or --match is required")
	}

	config.Init("config", ".")

	var cfg = config.MustReadConfig()
	repo := repository.NewRepository(cfg)

	if path != "" {
		importSpordleCSV(repo.DB, path)
	}
	if matchOnly {
		matchKmasterVenues(repo.DB, force, dryRun, report)
	}
}

func importSpordleCSV(db *gorm.DB, path string) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer f.Close()
	r := csv.NewReader(f)

	if _, err := r.Read(); err != nil {
		log.Fatalf("failed to read header: %v", err)
	}

	var count int
	err = db.Transaction(func(tx *gorm.DB) error {
		for {
			line, err := r.Read()
			if err != nil {
				break
			}

			if len(line) < 19 {
				log.Printf("skipping row with %d columns", len(line))
				continue
			}

			record := SpordleSurface{
				VenueID:           strings.TrimSpace(line[0]),
				VenueName:         strings.TrimSpace(line[1]),
				VenueAddress:      strings.TrimSpace(line[2]),
				VenueCity:         strings.TrimSpace(line[3]),
				VenueRegion:       strings.TrimSpace(line[4]),
				VenueCountry:      strings.TrimSpace(line[5]),
				VenueAlias:        strings.TrimSpace(line[6]),
				SurfaceID:         parseUint(strings.TrimSpace(line[10])),
				SurfaceName:       strings.TrimSpace(line[11]),
				SurfaceAlias:      strings.TrimSpace(line[12]),
				SurfaceSports:     strings.TrimSpace(line[13]),
				SurfaceTimeZone:   strings.TrimSpace(line[14]),
				SurfaceType:       strings.TrimSpace(line[15]),
				SurfaceSize:       strings.TrimSpace(line[16]),
				LivebarnSurfaceID: strings.TrimSpace(line[17]),
			}

			if strings.TrimSpace(line[7]) != "" {
				v, err := strconv.ParseFloat(strings.TrimSpace(line[7]), 64)
				if err == nil {
					record.VenueLatitude = v
				}
			}
			if strings.TrimSpace(line[8]) != "" {
				v, err := strconv.ParseFloat(strings.TrimSpace(line[8]), 64)
				if err == nil {
					record.VenueLongitude = v
				}
			}
			record.VenuePostalCode = strings.TrimSpace(line[9])
			if strings.TrimSpace(line[18]) != "" {
				v, err := strconv.Atoi(strings.TrimSpace(line[18]))
				if err == nil {
					record.NumberOfGamesComing = v
				}
			}

			if err := tx.Save(&record).Error; err != nil {
				return err
			}
			count++
		}
		return nil
	})

	if err != nil {
		log.Fatalf("import failed: %v", err)
	}

	log.Printf("imported %d records", count)
}

func matchKmasterVenues(db *gorm.DB, force, dryRun bool, reportPath string) {
	for _, q := range provinceFixes {
		if err := db.Exec(q).Error; err != nil {
			log.Fatalf("province fix failed: %v", err)
		}
	}
	log.Printf("applied %d spordle province fixes", len(provinceFixes))

	var spordleRows []struct {
		VenueID      string
		VenueName    string
		VenueAddress string
		VenueCity    string
		VenueRegion  string
	}
	err := db.Raw(`SELECT venue_id, MIN(venue_name) AS venue_name, MIN(venue_address) AS venue_address,
		MIN(venue_city) AS venue_city, MIN(venue_region) AS venue_region
		FROM spordle_surfaces GROUP BY venue_id`).Scan(&spordleRows).Error
	if err != nil {
		log.Fatalf("failed to load spordle venues: %v", err)
	}

	byProvince := make(map[string][]spordleVenue)
	for _, r := range spordleRows {
		p := normalizeProvince(r.VenueRegion)
		if p == "" {
			continue
		}
		byProvince[p] = append(byProvince[p], spordleVenue{
			VenueID: r.VenueID,
			Name:    r.VenueName,
			Address: r.VenueAddress,
			City:    r.VenueCity,
			Region:  p,
		})
	}

	query := db.Model(&KmasterVenue{})
	if !force {
		query = query.Where("spordle_venue_id IS NULL")
	}
	var kmasters []KmasterVenue
	if err := query.Find(&kmasters).Error; err != nil {
		log.Fatalf("failed to load kmaster venues: %v", err)
	}

	var updates []kmasterUpdate
	var report []reportRow
	var unmatched, ambiguous, skipped int
	for _, k := range kmasters {
		if strings.TrimSpace(k.RinkAddress) == "" {
			skipped++
			report = append(report, newReportRow(k, "skipped: missing address"))
			continue
		}
		if strings.TrimSpace(k.City) == "" {
			skipped++
			report = append(report, newReportRow(k, "skipped: missing city"))
			continue
		}

		p := normalizeProvince(k.ProvinceState)
		candidates := byProvince[p]
		if len(candidates) == 0 {
			unmatched++
			report = append(report, newReportRow(k, "unmatched: no spordle venue in province"))
			continue
		}

		kmasterCity := cityTokens(k.City)
		kmasterAddr := addressTokens(k.RinkAddress)
		if len(kmasterCity) == 0 || len(kmasterAddr) == 0 {
			unmatched++
			report = append(report, newReportRow(k, "unmatched: no candidate above similarity thresholds"))
			continue
		}

		var matches []spordleVenue
		for _, s := range candidates {
			if strings.TrimSpace(s.Address) == "" || strings.TrimSpace(s.City) == "" {
				continue
			}
			if tokenSimilarity(kmasterCity, cityTokens(s.City)) < cityThreshold {
				continue
			}
			if tokenSimilarity(kmasterAddr, addressTokens(s.Address)) < addressThreshold {
				continue
			}
			matches = append(matches, s)
		}

		switch {
		case len(matches) == 0:
			unmatched++
			report = append(report, newReportRow(k, "unmatched: no candidate above similarity thresholds"))
		case len(matches) > 1:
			ambiguous++
			report = append(report, newReportRow(k, fmt.Sprintf("ambiguous: %d matching candidates", len(matches))))
		default:
			updates = append(updates, kmasterUpdate{ID: k.ID, VenueID: matches[0].VenueID})
		}
	}

	if reportPath != "" {
		if err := writeReport(reportPath, report); err != nil {
			log.Fatalf("failed to write report: %v", err)
		}
		log.Printf("report written to %s (%d records)", reportPath, len(report))
	}

	if dryRun {
		for _, u := range updates {
			log.Printf("[dry-run] would set kmaster id=%d spordle_venue_id=%s", u.ID, u.VenueID)
		}
	} else if len(updates) > 0 {
		err = db.Transaction(func(tx *gorm.DB) error {
			for _, u := range updates {
				if err := tx.Model(&KmasterVenue{}).Where("id = ?", u.ID).
					Update("spordle_venue_id", u.VenueID).Error; err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			log.Fatalf("failed to update kmaster venues: %v", err)
		}
	}

	log.Printf("matching: processed=%d matched=%d unmatched=%d ambiguous=%d skipped(no address/city)=%d",
		len(kmasters), len(updates), unmatched, ambiguous, skipped)
}

func newReportRow(k KmasterVenue, reason string) reportRow {
	return reportRow{
		ID:            k.ID,
		VenueName:     k.VenueName,
		RinkAddress:   k.RinkAddress,
		City:          k.City,
		ProvinceState: k.ProvinceState,
		Reason:        reason,
	}
}

func writeReport(path string, rows []reportRow) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write([]string{"id", "venue_name", "rink_address", "city", "province_state", "reason"}); err != nil {
		return err
	}
	for _, r := range rows {
		if err := w.Write([]string{
			strconv.FormatUint(r.ID, 10),
			r.VenueName,
			r.RinkAddress,
			r.City,
			r.ProvinceState,
			r.Reason,
		}); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func normalizeProvince(s string) string {
	return strings.ToUpper(strings.TrimSpace(stripAccents(s)))
}

func normalize(s string) string {
	s = stripAccents(strings.ToLower(s))
	var b strings.Builder
	prevSpace := true
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevSpace = false
		} else if !prevSpace {
			b.WriteByte(' ')
			prevSpace = true
		}
	}
	return strings.TrimSpace(b.String())
}

func stripAccents(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	out, _, err := transform.String(t, s)
	if err != nil {
		return s
	}
	return out
}

func tokens(s string, expand map[string]string, stop map[string]bool) []string {
	fields := strings.Fields(normalize(s))
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if e, ok := expand[f]; ok {
			f = e
		}
		if stop[f] {
			continue
		}
		out = append(out, f)
	}
	return out
}

func cityTokens(s string) []string {
	return tokens(s, cityExpand, cityStopwords)
}

func addressTokens(s string) []string {
	return tokens(s, streetSuffixes, addressStopwords)
}

func tokenSimilarity(a, b []string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	setA := make(map[string]bool, len(a))
	for _, t := range a {
		setA[t] = true
	}
	common := 0
	seen := make(map[string]bool)
	for _, t := range b {
		if setA[t] && !seen[t] {
			common++
			seen[t] = true
		}
	}
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	return float64(common) / float64(minLen)
}

func parseUint(s string) uint {
	if s == "" {
		return 0
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return uint(v)
}
