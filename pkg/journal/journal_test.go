package journal

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/util"
	"testing"
)

func TestParseUserJournalLine(t *testing.T) {
	simpleData := []struct {
		input string
		user  User
	}{
		{"Hello\tWorld", User{Name: "Hello", Address: "World"}},
		{"  Spacey  \t  Address  ", User{Name: "  Spacey  ", Address: "  Address  "}},
		{"\t", User{Name: "", Address: ""}},
	}

	for _, entry := range simpleData {
		user, err := ParseUserJournalLine(entry.input)
		if assert.NoErrorf(t, err, "\"%s\" should be parsed without errors", entry.input) {
			assert.Equal(t, entry.user, user, "failed to parse line correctly")
		}
	}

	errorData := []string{
		"Hello World", "Hello\tWorld\t!",
	}

	for _, entry := range errorData {
		_, err := ParseUserJournalLine(entry)
		assert.Errorf(t, err, "\"%s\" does not have the correct amount of tab separators and should fail", entry)
	}
}

func TestUser_ToJournalLine(t *testing.T) {
	data := []struct {
		user User
		line string
	}{
		{User{Name: "Frank", Address: "Leipzig"}, "Frank\tLeipzig"},
		{User{Name: "", Address: ""}, "\t"},
		{User{Name: "\t", Address: "\t"}, "    \t    "},
	}

	for _, entry := range data {
		assert.Equalf(t, entry.line, entry.user.ToJournalLine(), "user not serialized correctly: %#v", entry.user)
	}
}

func TestParseEventJournalEntry(t *testing.T) {
	users := make(map[string]*User, 10)
	Locations = map[string]*Location{
		"MOS": {Name: "Mosbach", Code: "MOS"},
		"TST": {Name: "Test", Code: "TST"},
	}
	hash1, user1 := AddUserEntry(users, &User{Name: "Frank", Address: "Leipzig"})
	hash2, user2 := AddUserEntry(users, &User{Name: "Hello", Address: "World"})
	validData := []struct {
		hash  []byte
		event Event
	}{
		{
			hash:  hash1,
			event: Event{EventType: LOGIN, Location: Locations["MOS"], User: user1, Timestamp: 1609455600},
		},
		{
			hash:  hash1,
			event: Event{EventType: LOGOUT, Location: Locations["TST"], User: user1, Timestamp: 1634112969},
		},
		{
			hash:  hash2,
			event: Event{EventType: LOGIN, Location: Locations["TST"], User: user2, Timestamp: 0},
		},
	}

	for _, entry := range validData {
		data := fmt.Sprintf("%s\t%s\t%d", util.Base64Encode(entry.hash), entry.event.Location.Code, entry.event.Timestamp)
		event, err := ParseEventJournalEntry(entry.event.EventType, data, &users)
		if assert.NoErrorf(t, err, "failed to parse correct journal entry with %v and %s", entry.event.EventType, data) {
			assert.Equal(t, entry.event, event, "failed to correctly parse journal entry")
		}
	}

	errorData := []struct {
		eventType EventType
		data      string
		message   string
	}{
		{LOGIN, ".\ti0\tTST", "parsing invalid base64 hash should fail"},
		{LOGIN, "\t0\tTST", "parsing empty user base64 hash should fail"},
		{LOGIN, util.Base64Encode(hash1) + "\t\t", "parsing an empty timestamp should fail"},
		{LOGIN, util.Base64Encode(hash1) + "\te\t", "parsing an invalid timestamp should fail"},
		{LOGIN, util.Base64Encode(hash1) + "\t0\tTST\ttest", "too many fields should fail"},
		{LOGIN, util.Base64Encode(hash1) + "\t0\tXYZ", "unknown location should fail"},
		{LOGIN, util.Base64Encode(hash1) + "\t0", "not enough fields should fail"},
		{LOGIN, util.Base64Encode([]byte("12345678901234567890")) + "\t0\tTST", "parsing an unknown user hash should fail"},
	}

	for _, entry := range errorData {
		_, err := ParseEventJournalEntry(entry.eventType, entry.data, &users)
		assert.Errorf(t, err, "%s - data: %s", entry.message, entry.data)
	}
}

func AddUserEntry(users map[string]*User, user *User) ([]byte, *User) {
	hash := util.Hash(user)
	users[string(hash)] = user
	return hash, user
}
