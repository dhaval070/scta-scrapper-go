package main

import (
	"log"
	"testing"

	"github.com/antchfx/htmlquery"
	"github.com/stretchr/testify/assert"
)

func TestParseTime(t *testing.T) {
	html := `  <div class="cell small-2 text-center">
      <div class="time-primary">5:30 PM</div>
   </div>`
	result := parseTime(html)

	assert.Equal(t, "17:30", result)
}

func TestParseSchedules(t *testing.T) {
	doc, err := htmlquery.LoadDoc("/home/dhaval/scta.html")

	if err != nil {
		log.Fatal(err)
	}

	expected := [][]string{
		{
			"2022-10-17 19:15",
			"U10",
			"Hamilton Jr Bulldogs",
			"Halton Hurricanes",
			"Sherwood (1)",
		},
		{
			"2022-10-17 19:30",
			"U14",
			"Southern Tier Admirals",
			"Burlington Eagles",
			"Appleby (1)",
		},
		{
			"2022-10-17 19:40",
			"U14",
			"Credit River Capitals",
			"Hamilton Jr Bulldogs",
			"Mohawk (Tim Horton's)",
		},
	}

	result := parseSchedules(doc, 20221017)

	assert.Equal(t, expected, result[:3])

}
