package repository

import (
	"calendar-scrapper/dao/model"
	"calendar-scrapper/pkg/entity"
	"encoding/json"
	"log"

	"gorm.io/gorm"
)

func (r *Repository) MasterImportJson(js []entity.JsonLocation) {
	r.db.Transaction(func(tx *gorm.DB) (err error) {
		for _, loc := range js {
			log.Println(loc.ID)
			err = r.MasterImportLoc(tx, loc)
			if err != nil {
				return err
			}
		}
		return err
	})
}

func (r *Repository) MasterImportLoc(db *gorm.DB, l entity.JsonLocation) error {
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
		TotalSurfaces:       int32(len(l.Surfaces)),
	}
	b, _ := json.Marshal(l.LogoURL)
	loc.LogoURL = string(b)

	province := model.Province{ID: l.Province.ID}

	if l.Province.Name != "Ontario" {
		return nil
	}

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
		err = r.MasterImportSurface(db, l.ID, surface)
		if err != nil {
			return err
		}
	}

	return err
}

func (r *Repository) MasterImportSurface(db *gorm.DB, locId int32, surface entity.JsonSurface) (err error) {
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

	s.FirstMediaDate = surface.FirstMedia.FirstMediaDate
	if err = db.Save(&s).Error; err != nil {
		return err
	}

	for _, fm := range surface.FeedModes {
		if err = db.Save(&model.FeedMode{
			ID:       fm.ID,
			FeedMode: fm.Name,
		}).Error; err != nil {
			return err
		}

		db.Save(&model.SurfaceFeedMode{
			SurfaceID:  surface.ID,
			FeedModeID: fm.ID,
		})
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

	return err
}
