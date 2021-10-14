package journal

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/util"
	"log"
	"os"
	"path"
	"sync"
	"testing"
	"time"
)

func TestGetCurrentJournalPath(t *testing.T) {
	t.Parallel()
	journalPath := GetCurrentJournalPath("")
	now := time.Now()
	assert.Equal(t, fmt.Sprintf("%04d%02d%02d.txt", now.Year(), now.Month(), now.Day()), journalPath, "journal path determined incorrectly")
}

func TestNewWriter(t *testing.T) {
	t.Parallel()
	tempDir, remover := CreateTempDir(t)
	defer remover()
	writer, err := NewWriter(tempDir)
	if assert.NoError(t, err, "failed to create journal writer for new directory") {
		file, ok := writer.output.(*os.File)
		require.True(t, ok, "failed to dereference journal output to file")
		assert.Equal(t, GetCurrentJournalPath(tempDir), file.Name(), "wrong file output path")
	}
	file, ok := writer.output.(io.Closer)
	if ok {
		_ = file.Close()
	}

	_ = os.Remove(GetCurrentJournalPath(tempDir))
	_ = os.Mkdir(GetCurrentJournalPath(tempDir), 0777)
	writer, err = NewWriter(tempDir)
	assert.Error(t, err, "writer creation should fail if no output file can be created")
}

func TestWriter_LoadFrom(t *testing.T) {
	t.Parallel()
	tempDir, remover := CreateTempDir(t)
	defer remover()
	filePath := GetCurrentJournalPath(tempDir)

	writer := Writer{
		knownUsers: util.NewStringSet(10),
		outputLock: sync.Mutex{},
		directory:  tempDir,
	}
	assert.Error(t, writer.LoadFrom(filePath), "LoadFrom should not be able to read from non-existing files")

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	require.NoError(t, err, "internal error: failed to create journal file")

	users := []*User{
		{Name: "abc", Address: "def"},
		{Name: "cde", Address: "123"},
		{Name: "", Address: ""},
	}
	err = util.WriteString(file, "+ignored\n-ignored\n")
	for _, user := range users {
		err = util.WriteString(file, fmt.Sprintf("*%s\t%s\n", user.Name, user.Address))
	}
	err = file.Close()

	if assert.NoError(t, writer.LoadFrom(filePath), "failed to load writer data from prepared file") {
		for _, user := range users {
			assert.Truef(t, writer.knownUsers.Contains(string(util.Hash(user))), "missing user: %#v", user)
		}
		assert.LessOrEqual(t, writer.knownUsers.Size(), len(users), "too many entries in knownUsers")
	}

	file, err = os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	require.NoError(t, err, "internal error: failed to create journal file")
	err = util.WriteString(file, "*invalid line")
	err = file.Close()

	buf := bytes.Buffer{}
	reset := LogToBuffer(&buf)
	defer reset()
	assert.NoError(t, writer.LoadFrom(filePath), "writer load didn't fail on invalid line")
	reset()
	line := buf.String()
	if assert.NoError(t, err, "failed to parse log for incorrect user line") {
		assert.Equal(t, "Failed to parse user line \"invalid line\"\n", line, "incorrect error log for user line")
	}
}

func TestWriter_UpdateOutput(t *testing.T) {
	t.Parallel()
	tempDir, remover := CreateTempDir(t)
	defer remover()
	writer := Writer{outputLock: sync.Mutex{}, directory: tempDir}

	if assert.NoError(t, writer.UpdateOutput(), "failed to run update output") {
		file, ok := writer.output.(*os.File)
		require.True(t, ok, "failed to dereference journal output to file")
		if assert.Equal(t, GetCurrentJournalPath(tempDir), file.Name(), "incorrect journal file path") {
			assert.FileExists(t, file.Name(), "failed to create journal file")
		}
	}

	writer.directory = path.Join(writer.directory, "none")
	if assert.NoError(t, writer.UpdateOutput(), "failed to run update output for non-existing directory") {
		file, ok := writer.output.(*os.File)
		require.True(t, ok, "failed to dereference journal output to file")
		if assert.Equal(t, GetCurrentJournalPath(writer.directory), file.Name(), "incorrect journal file name") {
			assert.FileExists(t, file.Name(), "failed to create journal file")
		}
	}

	file, _ := os.OpenFile(path.Join(tempDir, "some"), os.O_CREATE|os.O_WRONLY, 0777)
	_ = file.Close()
	writer.directory = path.Join(tempDir, "some")
	assert.Error(t, writer.UpdateOutput(), "failing to create directories should be reported")

	writer.directory = path.Join(tempDir, "exist")
	_ = os.MkdirAll(GetCurrentJournalPath(writer.directory), 0777)
	assert.Error(t, writer.UpdateOutput(), "failing to create the journal output should be reported")
}

func TestWriter_writeLine(t *testing.T) {
	t.Parallel()
	buffer := &bytes.Buffer{}
	writer := Writer{
		outputLock: sync.Mutex{},
		output:     buffer,
	}
	if assert.NoError(t, writer.writeLine("test"), "valid write line failed") {
		assert.Equal(t, "test\n", buffer.String())
	}

	ew := newErrorWriter()
	writer.output = &ew
	assert.EqualError(t, writer.writeLine("test"), "failed to write journal line: test error")
}

func TestWriter_WriteUserIfUnknown(t *testing.T) {
	t.Parallel()
	buffer := &bytes.Buffer{}
	writer := Writer{
		knownUsers: util.NewStringSet(1),
		outputLock: sync.Mutex{},
		output:     buffer,
	}

	user := User{Name: "Tester", Address: "Addr"}
	hash := util.Base64Encode(util.Hash(user))
	retHash, err := writer.WriteUserIfUnknown(&user)
	assert.Equal(t, hash, retHash, "the returned hash should be accurate")
	if assert.NoError(t, err) {
		assert.Equal(t, fmt.Sprintf("*%s\t%s\n", user.Name, user.Address), buffer.String())
	}

	buffer.Reset()
	writer.knownUsers = util.NewStringSet(1)
	writer.knownUsers.Add(hash)
	retHash, err = writer.WriteUserIfUnknown(&user)
	assert.Equal(t, hash, retHash, "the returned hash should be accurate")
	if assert.NoError(t, err) {
		assert.Equal(t, "", buffer.String())
	}

	buffer.Reset()
	writer.knownUsers = util.NewStringSet(0)
	ew := newErrorWriter()
	writer.output = &ew
	_, err = writer.WriteUserIfUnknown(&user)
	if assert.Error(t, err, "errors in the writer should be propagated") {
		assert.Equal(t, "", buffer.String(), "errors should produce no output!")
	}
}

func TestWriter_WriteEventUserHash(t *testing.T) {
	t.Parallel()
	loc1 := &Location{Name: "Mosbach", Code: "MOS"}
	loc2 := &Location{Name: "Teststadt", Code: "TST"}
	buffer := &bytes.Buffer{}
	writer := Writer{
		knownUsers: util.NewStringSet(2),
		outputLock: sync.Mutex{},
		output:     buffer,
	}
	data := []struct {
		Hash string
		*Location
		EventType
		Result string
	}{
		{"hash1", loc1, LOGIN, "+hash1\tMOS"},
		{"hash1", loc1, LOGOUT, "-hash1\tMOS"},
		{"hash1", loc2, LOGIN, "+hash1\tTST"},
		{"hash2", loc1, LOGIN, "+hash2\tMOS"},
		{"hash2", loc1, LOGOUT, "-hash2\tMOS"},
		{"hash2", loc2, LOGOUT, "-hash2\tTST"},
	}

	for _, entry := range data {
		buffer.Reset()
		if assert.NoError(t, writer.WriteEventUserHash(entry.Hash, entry.Location, entry.EventType)) {
			assert.Equal(t, fmt.Sprintf("%s\t%d\n", entry.Result, time.Now().Unix()), buffer.String())
		}
	}

	ew := newErrorWriter()
	writer.output = &ew
	assert.Error(t, writer.WriteEventUserHash("hash", loc1, LOGIN))
}

func TestWriter_WriteEventUser(t *testing.T) {
	t.Parallel()
	loc1 := &Location{Name: "Mosbach", Code: "MOS"}
	loc2 := &Location{Name: "Teststadt", Code: "TST"}
	buffer := bytes.Buffer{}
	writer := Writer{
		knownUsers: util.NewStringSet(2),
		outputLock: sync.Mutex{},
		output:     &buffer,
	}
	user1 := User{Name: "Tester", Address: "TAddr"}
	hash1 := util.Base64Encode(util.Hash(user1))
	user2 := User{Name: "", Address: ""}
	hash2 := util.Base64Encode(util.Hash(user2))

	if assert.NoError(t, writer.WriteEventUser(&user1, loc1, LOGIN)) {
		assert.Equal(
			t, fmt.Sprintf("*%s\t%s\n+%s\tMOS\t%d\n", user1.Name, user1.Address, hash1, time.Now().Unix()),
			buffer.String(),
		)
	}

	buffer.Reset()
	writer.knownUsers.Add(hash2)
	if assert.NoError(t, writer.WriteEventUser(&user2, loc2, LOGOUT)) {
		assert.Equal(
			t, fmt.Sprintf("-%s\tTST\t%d\n", hash2, time.Now().Unix()),
			buffer.String(),
		)
	}

	buffer.Reset()
	ew := newErrorWriter()
	writer.output = &ew
	assert.Error(t, writer.WriteEventUser(&user2, loc1, LOGIN))

	writer.knownUsers.Remove(hash1)
	assert.Error(t, writer.WriteEventUser(&user1, loc1, LOGOUT))
}

func CreateTempDir(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err, "internal error: failed to create temp dir")
	return tempDir, func() {
		_ = os.Remove(tempDir)
	}
}

func LogToBuffer(buffer *bytes.Buffer) func() {
	log.SetOutput(buffer)
	flags := log.Flags()
	log.SetFlags(0)

	return func() {
		log.SetOutput(os.Stderr)
		log.SetFlags(flags)
	}
}

type errorWriter bool

func newErrorWriter() errorWriter {
	return false
}

func (ew *errorWriter) Write(_ []byte) (int, error) {
	return 0, fmt.Errorf("test error")
}
