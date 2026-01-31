package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/antchfx/htmlquery"
	"github.com/stretchr/testify/assert"
)

func TestParseVenue(t *testing.T) {
	b, err := os.ReadFile("../../var/mhr-livebarn-loc.html")
	if err != nil {
		t.Fatal(err)
	}

	doc, err := htmlquery.Parse(bytes.NewReader(b))
	if err != nil {
		t.Fatalf("failed to parse doc %v\n", err)
		return
	}

	res, err := parse_venue(doc)
	if err != nil {
		t.Fatalf("failed %v", err)
	}

	phone := "208-634-3570"
	var aka *string = nil
	var notes *string = nil
	var streaming = "https://livebarn.com"
	var website = "https://manchestericecenter.com/"

	expected := MhrLocation{
		RinkName:          "Manchester Ice & Events Centre",
		Aka:               aka,
		Address:           "200 E Lake St McCall, ID 83638",
		Streaming:         &streaming,
		Website:           &website,
		Phone:             &phone,
		Notes:             notes,
		LivebarnInstalled: true,
	}

	assert.Equal(t, expected, *res)
}
