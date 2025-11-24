package parser1

import (
	"testing"

	"github.com/antchfx/htmlquery"
	"github.com/stretchr/testify/assert"
)

func TestParseTournament(t *testing.T) {

	doc, err := htmlquery.LoadDoc("../../testdata/tournament-details.html")
	assert.NoError(t, err)

	result := ParseTournament(doc, "5636")
	assert.NotNil(t, result)
}
