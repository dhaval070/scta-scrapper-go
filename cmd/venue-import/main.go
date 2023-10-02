package main

import (
	"calendar-scrapper/config"
	"calendar-scrapper/dao/model"
	"encoding/json"
	"flag"
	"io"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func main() {
	var path string
	var err error
	flag.StringVar(&path, "path", "", "--path=<file path>")
	flag.Parse()

	if path == "" {
		panic("path is required")
	}

	config.Init("config", ".")

	cfg := config.MustReadConfig()
	db, err := gorm.Open(mysql.Open(cfg.DbDSN))
	if err != nil {
		panic(err)
	}
	// l := model.Location{}

	if err != nil {
		panic(err)
	}

	var js = []jsonLocation{}

	fh, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	data, err := io.ReadAll(fh)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(data, &js)
	if err != nil {
		panic(err)
	}
	// log.Printf("%#v\n", js[0].Surfaces)

	db.Transaction(func(tx *gorm.DB) (err error) {
		for _, loc := range js {
			err = importLoc(tx, loc)
			if err != nil {
				return err
			}
		}
		return err
	})
}

func genModels(db *gorm.DB) {
	g := gen.NewGenerator(gen.Config{
		OutPath: "./query",
		Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface, // generate mode
	})

	g.UseDB(db)
	g.ApplyBasic(g.GenerateAllTable()...)

	g.Execute()
}

func importLoc(db *gorm.DB, l jsonLocation) error {
	var err error
	loc := model.Location{
		ID:                  l.ID,
		Address1:            l.Address1,
		Address2:            l.Address2,
		City:                l.City,
		Name:                l.Name,
		UUID:                l.UUID,
		RecordingHoursLocal: l.RecordingHoursLocal,
		PostalCode:          l.PostalCode,
		AllSheetsCount:      l.AllSheetsCount,
		Longitude:           l.Longitude,
		Latitude:            l.Latitude,
	}
	b, _ := json.Marshal(l.LogoURL)
	loc.LogoURL = string(b)

	province := model.Province{ID: l.Province.ID}

	if err = db.First(&province).Error; err != nil {
		province.ProvinceName = l.Province.Name
		if err = db.Save(&province).Error; err != nil {
			return err
		}
	}
	loc.ProvinceID = l.Province.ID
	loc.VenueStatus = l.VenueStatus["name"].(string)
	loc.Zone = l.Zone["name"].(string)

	if err = db.Save(&loc).Error; err != nil {
		return err
	}

	for _, surface := range l.Surfaces {
		err = importSurface(db, l.ID, surface)
		if err != nil {
			return err
		}
	}

	return err
}

func importSurface(db *gorm.DB, locId int32, surface jsonSurface) (err error) {
	s := model.Surface{
		ID:         surface.ID,
		LocationID: locId,
		Name:       surface.Name,
		UUID:       surface.UUID,
		OrderIndex: surface.OrderIndex,
		VenueID:    surface.VenueID,
		ClosedFrom: surface.ClosedFrom,
		ComingSoon: surface.ComingSoon,
		Online:     surface.Online,
	}
	s.Status = surface.Status["name"].(string)

	if len(surface.Sports) > 0 {
		s.Sports = surface.Sports[0]["name"].(string)
	}

	for _, fm := range surface.FeedModes {
		if err = db.Save(&model.FeedMode{
			ID:       fm.ID,
			FeedMode: fm.Name,
		}).Error; err != nil {
			return err
		}

		if err = db.Save(&model.SurfaceFeedMode{
			SurfaceID:  surface.ID,
			FeedModeID: fm.ID,
		}).Error; err != nil {
			return err
		}
	}
	for _, rendition := range surface.Renditions {
		if err = db.Save(&model.Rendition{
			ID:        int32(rendition.ID),
			SurfaceID: surface.ID,
			Name:      rendition.Name,
			Width:     int32(rendition.Width),
			Height:    int32(rendition.Height),
			Ratio:     rendition.Ratio,
			Bitrate:   rendition.Bitrate,
		}).Error; err != nil {
			return err
		}
	}

	s.FirstMediaDate = surface.FirstMedia.FirstMediaDate
	err = db.Save(&s).Error
	return err
}

type jsonLocation struct {
	ID                  int32             `json:"id"`
	Address1            string            `json:"address1"`
	Address2            string            `json:"address2"`
	City                string            `json:"city"`
	Name                string            `json:"name"`
	UUID                string            `json:"uuid"`
	RecordingHoursLocal string            `json:"recordingHoursLocal"`
	PostalCode          string            `json:"postalCode"`
	AllSheetsCount      int32             `json:"allSheetsCount"`
	Longitude           float32           `json:"longitude"`
	Latitude            float32           `json:"latitude"`
	LogoURL             map[string]string `json:"logoUrl"`
	Province            jsonProvince      `json:"province"`
	VenueStatus         map[string]any    `json:"venueStatus"`
	Zone                map[string]any    `json:"zoneIds"`
	Surfaces            []jsonSurface     `json:"surfaces"`
}

type jsonProvince struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type jsonSurfaceMedia struct {
	// SurfaceUuid string
	FirstMediaDate int64 `json:"firstMediaDate"`
}

type jsonFreeMode struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type jsonSurface struct {
	ID         int32            `json:"id"`
	Name       string           `json:"name"`
	UUID       string           `json:"uuid"`
	OrderIndex int32            `json:"orderIndex"`
	VenueID    int32            `json:"venueId"`
	ClosedFrom int64            `json:"closedFrom"`
	ComingSoon bool             `json:"comingSoon"`
	Online     bool             `json:"online"`
	Status     map[string]any   `json:"surfaceStatus"`
	Sports     []map[string]any `json:"sports"`
	FeedModes  []jsonFreeMode   `json:"feedModes"`
	FirstMedia jsonSurfaceMedia `json:"firstMedia"`
	Renditions []jsonRendition  `json:"renditions"`
}

type jsonRendition struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Ratio   string `json:"ratio"`
	Bitrate int64  `json:"bitrate"`
}
