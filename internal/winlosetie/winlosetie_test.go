package winlosetie

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRun(t *testing.T) {
	c := &MockClient{}

	type Rec struct {
		VendorId        int    `json:vendorId`
		GameIdInternal  int    `json:"gameIdInternal"`
		LinkStreamVideo string `json:"linkStreamVideo"`
		VideoStatus     bool   `json:"videoStatus"`
	}
	expected := []Rec{
		{
			VendorId:        75,
			GameIdInternal:  123,
			LinkStreamVideo: "https://livebarn.com/en/video/1234/2022-12-12/12:00",
			VideoStatus:     true,
		},
		{
			VendorId:        75,
			GameIdInternal:  223,
			LinkStreamVideo: "https://livebarn.com/en/video/4456/2000-11-02/12:12",
			VideoStatus:     true,
		},
	}

	c.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		r, err := req.GetBody()
		assert.NoError(t, err)

		b, err := io.ReadAll(r)
		assert.NoError(t, err)
		var actual []Rec
		err = json.Unmarshal(b, &actual)
		assert.NoError(t, err)
		if !cmp.Equal(expected, actual) {
			log.Println(cmp.Diff(expected, actual))
		}
		return cmp.Equal(expected, actual)
	})).Return(&http.Response{Status: "ok", StatusCode: 200}, nil)

	client = c

	data := `123,2022-12-12,12:00,1234
223,2000-11-02,12:12,4456`
	err := run(strings.NewReader(data))
	assert.NoError(t, err)
}
