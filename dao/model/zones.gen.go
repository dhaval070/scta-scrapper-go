// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

const TableNameZone = "zones"

// Zone mapped from table <zones>
type Zone struct {
	ID       int32  `gorm:"column:id;primaryKey" json:"id"`
	ZoneName string `gorm:"column:zone_name;not null" json:"zone_name"`
}

// TableName Zone's table name
func (*Zone) TableName() string {
	return TableNameZone
}