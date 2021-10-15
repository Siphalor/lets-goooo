package journal

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/util"
	"os"
	"path"
	"testing"
)

func TestReadJournal(t *testing.T) {
	tempDir, destroy := CreateTempDir(t)
	defer destroy()

	dirPath := path.Join(tempDir, "dir")
	err := os.Mkdir(dirPath, 0777)
	require.NoError(t, err, "internal error: failed to create test directory")
	_, err = ReadJournal(dirPath)
	assert.Error(t, err, "providing a directory should fail the journal read in")

	filepath := path.Join(tempDir, "journal.txt")
	file, _ := os.OpenFile(filepath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0777)

	Locations = map[string]*Location{
		"MOS": {Name: "Mosbach", Code: "MOS"},
		"TST": {Name: "Testbach", Code: "TST"},
	}

	user1 := User{Name: "JLA", Address: "Mosbach"}
	hash1 := util.Base64Encode(util.Hash(user1))
	user2 := User{Name: "Tester", Address: "Goland"}
	hash2 := util.Base64Encode(util.Hash(user2))

	_ = util.WriteString(file, fmt.Sprintf("*%s\t%s\n", user1.Name, user1.Address))
	_ = util.WriteString(file, fmt.Sprintf("+%s\tMOS\t0\n", hash1))
	_ = util.WriteString(file, fmt.Sprintf("*%s\t%s\n", user2.Name, user2.Address))
	_ = util.WriteString(file, fmt.Sprintf("+%s\tTST\t20\n", hash2))
	_ = util.WriteString(file, fmt.Sprintf("-%s\tMOS\t10\n", hash1))
	_ = util.WriteString(file, fmt.Sprintf("-%s\tTST\t30\n", hash2))
	_ = file.Close()

	journal, err := ReadJournal(filepath)
	if assert.NoError(t, err, "valid journal file failed reading") {
		assert.Equal(t, 2, len(journal.users), "incorrect number of users in journal")
		readUser1, exists := journal.users[string(util.Hash(user1))]
		require.True(t, exists, "readUser1 1 doesn't exist in journal")
		assert.Equal(t, user1, *readUser1, "readUser1 1 is read incorrectly")
		readUser2, exists := journal.users[string(util.Hash(user2))]
		require.True(t, exists, "readUser1 2 doesn't exist in journal")
		assert.Equal(t, user2, *readUser2, "readUser1 2 is read incorrectly")

		assert.Equal(t, []Event{
			{LOGIN, readUser1, Locations["MOS"], 0},
			{LOGIN, readUser2, Locations["TST"], 20},
			{LOGOUT, readUser1, Locations["MOS"], 10},
			{LOGOUT, readUser2, Locations["TST"], 30},
		}, journal.events, "events are read incorrectly")
	}
}

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
