package rockies

import "time"

type Game struct {
	ArenaName    string    `json:"ArenaName"`
	CategoryName string    `json:"CategoryName"`
	EDate        string    `json:"eDate"` // 2025-11-01T08:15:00
	Date         time.Time `json:"-"`
	HomeDivision string    `json:"HomeDivision"`
	HomeTeamName string    `json:"HomeTeamName"`
	AwayTeamName string    `json:"AwayTeamName"`
	Country      string    `json:"Country"`
	Prov         string    `json:"Prov"`
	RARIDString  string    `json:"RARIDString"`
	Address      string    `json:"-"`
}
