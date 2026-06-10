package model

import "time"

const TableNameApiKey = "api_keys"

type ApiKey struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	KeyHash   string    `gorm:"column:key_hash;uniqueIndex;size:64;not null" json:"key_hash"`
	Name      string    `gorm:"column:name;not null" json:"name"`
	Active    int8      `gorm:"column:active;not null;default:1" json:"active"`
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (*ApiKey) TableName() string {
	return TableNameApiKey
}
