package main

import (
	"log"
	"testing"

	"github.com/antchfx/htmlquery"
	"github.com/stretchr/testify/assert"
)

func TestParseGroups(t *testing.T) {
	var err error
	doc, err := htmlquery.LoadDoc("../../testdata/etahockey/groups.html")

	assert.NoError(t, err)

	result := parseGroups(doc)
	log.Println(result)
	expected := map[string]string{
		"U10": "1077",
		"U11": "1078",
		"U12": "1079",
		"U13": "1080",
		"U14": "1081",
		"U15": "1082",
		"U16": "1083",
		"U18": "1084",
	}
	assert.Equal(t, expected, result)
}
