package rockies

import "time"

type Game struct {
	AID          int       `json:"AID"`
	ArenaName    string    `json:"ArenaName"`
	CategoryName string    `json:"CategoryName"`
	SDate        string    `json:"sDate"`
	Date         time.Time `json:"-"`
	HomeDivision string    `json:"HomeDivision"`
	HomeTeamName string    `json:"HomeTeamName"`
	AwayTeamName string    `json:"AwayTeamName"`
	Country      string    `json:"Country"`
	Prov         string    `json:"Prov"`
	RARIDString  string    `json:"RARIDString"`
	Address      string    `json:"-"`
}
