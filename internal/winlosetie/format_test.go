package winlosetie

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatNyhl(t *testing.T) {
	input := `795847,North York Hockey League,Playoffs 2024,U11,TIER 2,GROUP B,Forest Hill,TIER 2,GROUP B,Mimico,Victoria Village,3/7/2024,6:10 PM,4426
796037,North York Hockey League,Playoffs 2024,U10,TIER 2,,York Mills,TIER 2,,Avenue Road,Downsview,3/8/2024,7:10 PM,4195`

	data, err := formatNyhl(strings.NewReader(input))
	assert.NoError(t, err)

	expected := []DataRec{
		{
			Id:        795847,
			Date:      "2024-03-07",
			Time:      "18:10",
			SurfaceID: "4426",
		},
		{
			Id:        796037,
			Date:      "2024-03-08",
			Time:      "19:10",
			SurfaceID: "4195",
		},
	}

	assert.Equal(t, expected, data)
}
