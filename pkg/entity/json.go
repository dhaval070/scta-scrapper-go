package entity

type JsonLocation struct {
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
	Province            JsonProvince      `json:"province"`
	VenueStatus         map[string]any    `json:"venueStatus"`
	Zone                map[string]any    `json:"zoneIds"`
	Surfaces            []JsonSurface     `json:"surfaces"`
}

type JsonProvince struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type JsonSurfaceMedia struct {
	// SurfaceUuid string
	FirstMediaDate int64 `json:"firstMediaDate"`
}

type JsonFreeMode struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type JsonSurface struct {
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
	FeedModes  []JsonFreeMode   `json:"feedModes"`
	FirstMedia JsonSurfaceMedia `json:"firstMedia"`
	Renditions []JsonRendition  `json:"renditions"`
}

type JsonRendition struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Ratio   string `json:"ratio"`
	Bitrate int64  `json:"bitrate"`
}
