package model

import "time"

const TableNameKmasterVenueList = "kmaster_venue_list"

type KmasterVenueList struct {
	ID                uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Validate          int8      `gorm:"column:validate;not null;default:0" json:"validate"`
	LivebarnVenueID   int       `gorm:"column:livebarn_venue_id" json:"livebarn_venue_id"`
	MhrVenueID        int       `gorm:"column:mhr_venue_id" json:"mhr_venue_id"`
	VenueName         string    `gorm:"column:venue_name;not null" json:"venue_name"`
	Surfaces          int       `gorm:"column:surfaces;not null;default:0" json:"surfaces"`
	City              string    `gorm:"column:city" json:"city"`
	RinkAddress       string    `gorm:"column:rink_address" json:"rink_address"`
	PostalCode        string    `gorm:"column:postal_code" json:"postal_code"`
	ProvinceState     string    `gorm:"column:province_state" json:"province_state"`
	Country           string    `gorm:"column:country" json:"country"`
	CompanyNameAlt1   string    `gorm:"column:company_name_alt1" json:"company_name_alt1"`
	CompanyNameAlt2   string    `gorm:"column:company_name_alt2" json:"company_name_alt2"`
	CompanyNameAlt3   string    `gorm:"column:company_name_alt3" json:"company_name_alt3"`
	ParentCompany     string    `gorm:"column:parent_company" json:"parent_company"`
	VenueType         string    `gorm:"column:venue_type" json:"venue_type"`
	AccountStatus     string    `gorm:"column:account_status" json:"account_status"`
	StreamingPlatform string    `gorm:"column:streaming_platform" json:"streaming_platform"`
	PhoneNumber       string    `gorm:"column:phone_number" json:"phone_number"`
	Website           string    `gorm:"column:website" json:"website"`
	Latitude          *float64  `gorm:"column:latitude" json:"latitude"`
	Longitude         *float64  `gorm:"column:longitude" json:"longitude"`
	CreatedAt         time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (*KmasterVenueList) TableName() string {
	return TableNameKmasterVenueList
}
