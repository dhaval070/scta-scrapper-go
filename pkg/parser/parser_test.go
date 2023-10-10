package parser

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
	result := ParseTime(html)

	assert.Equal(t, "17:30", result)
}

func TestParseSchedules(t *testing.T) {
	doc, err := htmlquery.LoadDoc("../../testdata/scta.html")

	if err != nil {
		log.Fatal(err)
	}

	expected := [][]string{
		{
			"2022-10-17 19:15",
			"",
			"Halton Hurricanes",
			"Hamilton Jr Bulldogs",
			"Sherwood (1)",
			"U10-116",
		},
		{
			"2022-10-17 19:30",
			"",
			"Burlington Eagles",
			"Southern Tier Admirals",
			"Appleby (1)",
			"U14-18",
		},
		{
			"2022-10-17 19:40",
			"",
			"Hamilton Jr Bulldogs",
			"Credit River Capitals",
			"Mohawk (Tim Horton's)",
			"U14-19",
		},
	}

	result := ParseSchedules(doc, 20221017)

	assert.Equal(t, expected, result[:3])

	doc, err = htmlquery.LoadDoc("../../testdata/etahockey/U11.html")

	if err != nil {
		log.Fatal(err)
	}

	result = ParseSchedules(doc, 20221017)
	log.Println(result)
	expected = [][]string{
		{
			"2022-10-19 19:00",
			"",
			"Kingston Jr Gaels",
			"Quinte Red Devils",
			"INVISTA (Desjardins)",
			"U11 - 052",
		},
	}

	assert.Equal(t, expected, result[:1])
}
