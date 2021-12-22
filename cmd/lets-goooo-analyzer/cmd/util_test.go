package cmd

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/internal/journal"
	"testing"
)

func TestFindUser(t *testing.T) {
	journal.Locations = map[string]*journal.Location{
		"TST": {Code: "TST", Name: "Teststadt"},
		"HST": {Code: "HST", Name: "Hauptstadt"},
	}

	j, err := journal.ReadJournal("testdata/journal.txt")
	require.NoError(t, err, "Failed to load journal from test data!")

	tester := journal.User{
		Name:    "Tester",
		Address: "Teststadt",
	}
	klaus := journal.User{
		Name:    "Klaus",
		Address: "Musterdorf",
	}

	if user, err := findUser(&j, "Tester", ""); assert.NoError(t, err) {
		assert.Equal(t, tester, *user)
	}
	if user, err := findUser(&j, "Klaus", ""); assert.NoError(t, err) {
		assert.Equal(t, klaus, *user)
	}
	if user, err := findUser(&j, "", "Teststadt"); assert.NoError(t, err) {
		assert.Equal(t, tester, *user)
	}
	if user, err := findUser(&j, "", "Musterdorf"); assert.NoError(t, err) {
		assert.Equal(t, klaus, *user)
	}
	if user, err := findUser(&j, "Tester", "Teststadt"); assert.NoError(t, err) {
		assert.Equal(t, tester, *user)
	}
	if user, err := findUser(&j, "Klaus", "Musterdorf"); assert.NoError(t, err) {
		assert.Equal(t, klaus, *user)
	}

	_, err = findUser(&j, "???", "")
	assert.Error(t, err)
	assert.Equal(t, 404, err.(*Error).Code())
	_, err = findUser(&j, "", "???")
	assert.Error(t, err)
	assert.Equal(t, 404, err.(*Error).Code())

	_, err = findUser(&j, "Tester", "Musterdorf")
	assert.Error(t, err)
	assert.Equal(t, 404, err.(*Error).Code())
	_, err = findUser(&j, "Klaus", "Teststadt")
	assert.Error(t, err)
	assert.Equal(t, 404, err.(*Error).Code())

	_, err = findUser(&j, "", "")
	assert.Error(t, err)
}
