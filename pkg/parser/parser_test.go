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
	var cases = map[string]struct {
		file     string
		date     int
		expected [][]string
	}{
		"scta": {
			file: "../../testdata/scta.html",
			expected: [][]string{
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
			},
			date: 20221017,
		},
		"beechey": {
			file: "../../testdata/beechey.html",
			expected: [][]string{
				{
					"2023-10-21 19:00",
					"",
					"Thorold Blackhawks",
					"West Niagara Flying Aces",
					"Thorold (Frank Doherty)",
					"U21069",
				},
			},
			date: 20221017,
		},
		"bluewaterhockey": {
			file: "../../testdata/bluewaterhockey.html",
			expected: [][]string{
				{
					"2023-10-14 13:00",
					"",
					"Windsor Jr. Spitfires",
					"Riverside Rangers",
					"Central Park (South)",
					"U11006",
				},
			},
			date: 20231014,
		},
	}

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			doc, err := htmlquery.LoadDoc(test.file)

			if err != nil {
				log.Fatal(err)
			}

			result := ParseSchedules(doc, test.date)
			assert.Equal(t, test.expected, result[:len(test.expected)])
		})
	}

	doc, err := htmlquery.LoadDoc("../../testdata/etahockey/U11.html")

	if err != nil {
		log.Fatal(err)
	}

	result := ParseSchedules(doc, 20221017)
	log.Println(result)
	expected := [][]string{
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
