package journal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJournal_GetUsers(t *testing.T) {
	emptyJournal := Journal{}
	for range emptyJournal.GetUsers() {
		assert.Fail(t, "empty journal is reporting users")
	}

	journal := Journal{users: map[string]*User{
		"1": {}, "2": {}, "3": {},
	}}
	encounters := make(map[*User]int, len(journal.users))
	for _, user := range journal.users {
		encounters[user] = 0
	}
	for user := range journal.GetUsers() {
		_, exists := encounters[user]
		if !exists {
			assert.Fail(t, "got unknown user %p")
		}
		encounters[user]++
	}
	for _, encs := range encounters {
		if encs < 1 {
			assert.Fail(t, "user hasn't occurred in iteration")
		} else if encs > 1 {
			assert.Fail(t, "user encountered multiple times in iteration")
		}
	}
}
