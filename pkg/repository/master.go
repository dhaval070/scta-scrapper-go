package repository

import (
	"calendar-scrapper/dao/model"
	"calendar-scrapper/pkg/entity"
	"encoding/json"
	"log"

	"gorm.io/gorm"
)

func (r *Repository) MasterImportJson(js []entity.JsonLocation) {
	batchSize := 100
	err := r.DB.Transaction(func(tx *gorm.DB) (err error) {
		provinces := make(map[int32]model.Province)
		locations := []model.Location{}
		surfaces := []model.Surface{}
		feedModes := make(map[int32]model.FeedMode)
		surfaceFeedModes := []model.SurfaceFeedMode{}
		renditions := []model.Rendition{}

		// Track IDs from JSON to identify deleted records
		jsonLocationIDs := make(map[int32]bool)
		jsonSurfaceIDs := make(map[int32]bool)

		for _, loc := range js {
			jsonLocationIDs[loc.ID] = true
			log.Println(loc.ID)

			// Collect provinces
			if _, exists := provinces[loc.Province.ID]; !exists {
				provinces[loc.Province.ID] = model.Province{
					ID:           loc.Province.ID,
					ProvinceName: loc.Province.Name,
				}
			}

			// Build location
			location := model.Location{
				ID:                  loc.ID,
				Address1:            loc.Address1,
				Address2:            loc.Address2,
				City:                loc.City,
				Name:                loc.Name,
				UUID:                loc.UUID,
				RecordingHoursLocal: loc.RecordingHoursLocal,
				PostalCode:          loc.PostalCode,
				AllSheetsCount:      loc.AllSheetsCount,
				Longitude:           loc.Longitude,
				Latitude:            loc.Latitude,
				TotalSurfaces:       int32(len(loc.Surfaces)),
				ProvinceID:          loc.Province.ID,
				VenueStatus:         loc.VenueStatus["name"].(string),
				Zone:                loc.Zone["name"].(string),
				DeletedAt:           nil,
			}
			b, _ := json.Marshal(loc.LogoURL)
			location.LogoURL = string(b)
			locations = append(locations, location)

			// Build surfaces
			for _, surface := range loc.Surfaces {
				jsonSurfaceIDs[surface.ID] = true
				s := model.Surface{
					ID:         surface.ID,
					LocationID: loc.ID,
					Name:       surface.Name,
					UUID:       surface.UUID,
					OrderIndex: surface.OrderIndex,
					VenueID:    surface.VenueID,
					ClosedFrom: surface.ClosedFrom,
					ComingSoon: surface.ComingSoon,
					Online:     surface.Online,
					Status:     surface.Status["name"].(string),
					DeletedAt:  nil,
				}
				if len(surface.Sports) > 0 {
					s.Sports = surface.Sports[0]["name"].(string)
				}
				s.FirstMediaDate = surface.FirstMedia.FirstMediaDate
				surfaces = append(surfaces, s)

				// Collect feed modes
				for _, fm := range surface.FeedModes {
					if _, exists := feedModes[fm.ID]; !exists {
						feedModes[fm.ID] = model.FeedMode{
							ID:       fm.ID,
							FeedMode: fm.Name,
						}
					}
					surfaceFeedModes = append(surfaceFeedModes, model.SurfaceFeedMode{
						SurfaceID:  surface.ID,
						FeedModeID: fm.ID,
					})
				}

				// Collect renditions
				for _, rendition := range surface.Renditions {
					renditions = append(renditions, model.Rendition{
						ID:        int32(rendition.ID),
						SurfaceID: surface.ID,
						Name:      rendition.Name,
						Width:     int32(rendition.Width),
						Height:    int32(rendition.Height),
						Ratio:     rendition.Ratio,
						Bitrate:   rendition.Bitrate,
					})
				}
			}
		}

		// Batch insert/update provinces
		provinceSlice := make([]model.Province, 0, len(provinces))
		for _, p := range provinces {
			provinceSlice = append(provinceSlice, p)
		}
		if len(provinceSlice) > 0 {
			if err = tx.Save(&provinceSlice).Error; err != nil {
				return err
			}
		}

		// Batch insert/update locations
		for i := 0; i < len(locations); i += batchSize {
			end := i + batchSize
			end = min(end, len(locations))
			batch := locations[i:end]
			if err = tx.Save(&batch).Error; err != nil {
				return err
			}
		}

		// Batch insert/update surfaces
		for i := 0; i < len(surfaces); i += batchSize {
			end := i + batchSize
			end = min(end, len(surfaces))
			batch := surfaces[i:end]
			if err = tx.Save(&batch).Error; err != nil {
				return err
			}
		}

		// Batch insert/update feed modes
		feedModeSlice := make([]model.FeedMode, 0, len(feedModes))
		for _, fm := range feedModes {
			feedModeSlice = append(feedModeSlice, fm)
		}
		if len(feedModeSlice) > 0 {
			if err = tx.Save(&feedModeSlice).Error; err != nil {
				return err
			}
		}

		// Batch insert/update surface feed modes
		for i := 0; i < len(surfaceFeedModes); i += batchSize {
			end := i + batchSize
			end = min(end, len(surfaceFeedModes))
			batch := surfaceFeedModes[i:end]
			if err = tx.Save(&batch).Error; err != nil {
				return err
			}
		}

		// Batch insert/update renditions
		for i := 0; i < len(renditions); i += batchSize {
			end := i + batchSize
			end = min(end, len(renditions))
			batch := renditions[i:end]
			if err = tx.Save(&batch).Error; err != nil {
				return err
			}
		}

		// Mark locations not in JSON as deleted
		var existingLocations []model.Location
		if err = tx.Where("deleted_at IS NULL").Find(&existingLocations).Error; err != nil {
			return err
		}
		deletedLocationIDs := []int32{}
		for _, loc := range existingLocations {
			if !jsonLocationIDs[loc.ID] {
				deletedLocationIDs = append(deletedLocationIDs, loc.ID)
			}
		}
		if len(deletedLocationIDs) > 0 {
			// Batch update deleted locations
			if err = tx.Model(&model.Location{}).Where("id IN ?", deletedLocationIDs).Update("deleted_at", gorm.Expr("NOW()")).Error; err != nil {
				return err
			}
			// Batch update surfaces of deleted locations
			if err = tx.Model(&model.Surface{}).Where("location_id IN ? AND deleted_at IS NULL", deletedLocationIDs).Update("deleted_at", gorm.Expr("NOW()")).Error; err != nil {
				return err
			}
		}

		// Mark surfaces not in JSON as deleted (for non-deleted locations)
		var existingSurfaces []model.Surface
		if err = tx.Where("deleted_at IS NULL").Find(&existingSurfaces).Error; err != nil {
			return err
		}
		deletedSurfaceIDs := []int32{}
		for _, surface := range existingSurfaces {
			if !jsonSurfaceIDs[surface.ID] {
				deletedSurfaceIDs = append(deletedSurfaceIDs, surface.ID)
			}
		}
		if len(deletedSurfaceIDs) > 0 {
			// Batch update deleted surfaces
			if err = tx.Model(&model.Surface{}).Where("id IN ?", deletedSurfaceIDs).Update("deleted_at", gorm.Expr("NOW()")).Error; err != nil {
				return err
			}
		}

		return err
	})
	if err != nil {
		panic(err.Error())
	}
}
