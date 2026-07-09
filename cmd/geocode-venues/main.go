package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"calendar-scrapper/config"
	"calendar-scrapper/pkg/repository"

	"gorm.io/gorm"
)

var errRateLimited = errors.New("rate limited")

type NominatimResult struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

type KmasterVenue struct {
	ID              uint64   `gorm:"column:id"`
	LivebarnVenueID *int     `gorm:"column:livebarn_venue_id"`
	RinkAddress     string   `gorm:"column:rink_address"`
	City            string   `gorm:"column:city"`
	ProvinceState   string   `gorm:"column:province_state"`
	PostalCode      string   `gorm:"column:postal_code"`
	Country         string   `gorm:"column:country"`
	Latitude        *float64 `gorm:"column:latitude"`
	Longitude       *float64 `gorm:"column:longitude"`
}

func (KmasterVenue) TableName() string {
	return "kmaster_venue_list"
}

func main() {
	config.Init("config", ".")
	cfg := config.MustReadConfig()
	repo := repository.NewRepository(cfg)
	db := repo.DB

	pass1(db)
	pass2(db)
}

func pass1(db *gorm.DB) {
	log.Println("Pass 1: matching livebarn_venue_id with locations table...")

	res := db.Exec(`UPDATE kmaster_venue_list k
		JOIN locations l ON k.livebarn_venue_id = l.id
		SET k.latitude = l.latitude, k.longitude = l.longitude
		WHERE k.latitude IS NULL`)

	if res.Error != nil {
		log.Fatalf("pass 1 failed: %v", res.Error)
	}
	log.Printf("pass 1: %d rows updated via livebarn_venue_id join", res.RowsAffected)
}

func pass2(db *gorm.DB) {
	log.Println("Pass 2: geocoding remaining venues via Nominatim...")

	var venues []KmasterVenue
	db.Where("latitude IS NULL").
		Where("rink_address != '' OR postal_code != ''").
		Find(&venues)

	if len(venues) == 0 {
		log.Println("pass 2: no venues to geocode")
		return
	}

	log.Printf("pass 2: geocoding %d venues", len(venues))

	client := &http.Client{Timeout: 15 * time.Second}
	var geocoded, failed int

	for i, v := range venues {
		params, countryCode := buildStructuredParams(v.RinkAddress, v.City, v.ProvinceState, v.PostalCode, v.Country)

		lat, lon, err := geocode(client, params, countryCode)
		if errors.Is(err, errRateLimited) {
			log.Printf("  [%d/%d] venue %d: rate limited (429), terminating", i+1, len(venues), v.ID)
			break
		}
		if err != nil {
			log.Printf("  [%d/%d] venue %d: geocode error: %v", i+1, len(venues), v.ID, err)
			failed++
		} else {
			log.Println("venue, lat, lon", v.ID, lat, lon)

			db.Model(&KmasterVenue{}).Where("id = ?", v.ID).
				Updates(map[string]any{"latitude": lat, "longitude": lon})
			geocoded++
		}

		if i < len(venues)-1 {
			time.Sleep(1100 * time.Millisecond)
		}
	}

	log.Printf("pass 2: geocoded %d, failed %d", geocoded, failed)
}

func buildStructuredParams(rinkAddress, city, provinceState, postalCode, country string) (url.Values, string) {
	params := url.Values{}
	if rinkAddress != "" {
		params.Set("street", rinkAddress)
	}
	if city != "" {
		params.Set("city", city)
	}
	if provinceState != "" {
		params.Set("state", provinceState)
	}
	if postalCode != "" {
		params.Set("postalcode", postalCode)
	}

	countryCode := ""
	if country != "" {
		countryCode = strings.ToLower(country)
	}

	return params, countryCode
}

func geocode(client *http.Client, params url.Values, countryCode string) (float64, float64, error) {
	params.Set("format", "json")
	params.Set("limit", "2")
	if countryCode != "" {
		if countryCode == "CA" {
			countryCode = "Canada"
		}
		params.Set("country", countryCode)
	}

	u := "https://nominatim.openstreetmap.org/search?" + params.Encode()

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:152.0) Gecko/20100101 Firefox/152.0")

	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return 0, 0, errRateLimited
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, 0, fmt.Errorf("http %d: %s, url: %v", resp.StatusCode, string(body), u)
	}

	var results []NominatimResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return 0, 0, fmt.Errorf("decoding response: %w", err)
	}

	if len(results) == 0 {
		return 0, 0, fmt.Errorf("no results for: %s", u)
	}

	if len(results) > 1 {
		b, err := json.Marshal(results)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to marshal json for multiple results")
		}
		log.Println(string(b))
		return 0, 0, fmt.Errorf("skipped: multiple results")
	}

	var lat, lon float64
	if _, err := fmt.Sscanf(results[0].Lat, "%f", &lat); err != nil {
		return 0, 0, fmt.Errorf("parsing lat %q: %w", results[0].Lat, err)
	}
	if _, err := fmt.Sscanf(results[0].Lon, "%f", &lon); err != nil {
		return 0, 0, fmt.Errorf("parsing lon %q: %w", results[0].Lon, err)
	}

	return lat, lon, nil
}
