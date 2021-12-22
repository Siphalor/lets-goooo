// Part of the Let's Goooo project
// Copyright 2021; matriculation numbers: 1103207, 3106445, 4485500
// Let's goooo get this over together

package journal

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/internal/util"
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
	tempDir := t.TempDir()
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

func TestNewWriter_existingJournal(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	filePath := GetCurrentJournalPath(tempDir)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0777)
	require.NoError(t, err, "internal error: failed to create test journal file")
	err = util.WriteString(file, "*Tester\tOrt\n")
	require.NoError(t, err, "internal error: failed to write to test journal file")
	_ = file.Close()

	writer, err := NewWriter(tempDir)
	require.NoError(t, err, "failed to read existing data")
	assert.Equal(t, map[string]*Location{"nPQeHgKWuAdyhGh6NPteN7LuDLg=": nil}, writer.knownUsers)
}

func TestWriter_LoadFrom(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()

	Locations = map[string]*Location{
		"TST": {Code: "TST", Name: "Teststadt"},
		"HST": {Code: "HST", Name: "Hauptstadt"},
	}

	writer := Writer{
		knownUsers: createKnownUserMap(10),
		outputLock: sync.Mutex{},
		directory:  tempDir,
	}
	assert.Error(t, writer.LoadFrom(path.Join(tempDir, "unkown")), "LoadFrom should not be able to read from non-existing files")

	if assert.NoError(t, writer.LoadFrom("testdata/import_journal.txt"), "failed to load writer data from prepared file") {
		assert.Equal(
			t,
			map[string]*Location{
				"P245C5uet9ZzSc0fXoOi7/0FB4I=": Locations["HST"],
				"ASkl/7Pm/MXnARb+f7+Fhk5GeYc=": nil,
				"oBklljrMPMa4Db3A4xsgTlfaLRw=": nil,
			}, writer.knownUsers,
		)
	}

	filePath := path.Join(tempDir, "invalid")
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
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
	tempDir := t.TempDir()
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

func TestUpdateOutput_complete(t *testing.T) {
	t.Parallel()

	expectedLog := "*Tester\tTeststadt\n" + "+HjLV+aPwKzq3szuae53Zv5n4puw=\tHST\t"

	tempDir := t.TempDir()
	oldLogPath := path.Join(tempDir, "old")
	oldFile, err := os.OpenFile(oldLogPath, os.O_CREATE|os.O_WRONLY, 0777)
	require.NoError(t, err, "internal error: failed to open new file for output")
	writer := Writer{
		knownUsers: createKnownUserMap(100),
		output:     oldFile,
		outputLock: sync.Mutex{},
		directory:  tempDir,
	}
	user := User{Name: "Tester", Address: "Teststadt"}
	location := Location{Name: "Hauptstadt", Code: "HST"}
	require.NoError(t, writer.WriteEventUser(&user, &location, LOGIN), "failed to write user event")
	content, err := ioutil.ReadFile(oldLogPath)
	require.NoError(t, err, "internal error: failed to read output file")
	assert.Contains(t, string(content), expectedLog, "invalid text written to original journal")

	require.NoError(t, writer.UpdateOutput(), "failed to update the output")
	require.NoError(t, writer.WriteEventUser(&user, &location, LOGIN), "failed to write user event")
	file, ok := writer.output.(*os.File)
	assert.False(t, file.Name() == oldFile.Name(), "the output file didn't change")
	require.True(t, ok, "output was not a file")
	content, err = ioutil.ReadFile(file.Name())
	require.NoError(t, err, "internal error: failed to read output file")
	assert.Contains(t, string(content), expectedLog, "when writing to a new file the user line should be printed again")
}

func TestWriter_GetCurrentUserLocation(t *testing.T) {
	loc1 := Location{Name: "Test1"}
	writer := Writer{
		knownUsers: map[string]*Location{
			"hash1": &loc1,
			"hash2": nil,
		},
	}
	loc, err := writer.GetCurrentUserLocation("hash1")
	if assert.NoError(t, err) {
		assert.Equal(t, &loc1, loc)
	}
	loc, err = writer.GetCurrentUserLocation("hash2")
	if assert.NoError(t, err) {
		assert.Nil(t, loc)
	}
	_, err = writer.GetCurrentUserLocation("unknown")
	assert.Error(t, err, "getting the location of an unknown user should produce an error")
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
		knownUsers: createKnownUserMap(1),
		outputLock: sync.Mutex{},
		output:     buffer,
	}

	user := User{Name: "Tester", Address: "Addr"}
	hash := util.Base64Encode(user.Hash())
	retHash, err := writer.WriteUserIfUnknown(&user)
	assert.Equal(t, hash, retHash, "the returned hash should be accurate")
	if assert.NoError(t, err) {
		assert.Equal(t, fmt.Sprintf("*%s\t%s\n", user.Name, user.Address), buffer.String())
	}

	buffer.Reset()
	writer.knownUsers = createKnownUserMap(1)
	writer.knownUsers[hash] = nil
	retHash, err = writer.WriteUserIfUnknown(&user)
	assert.Equal(t, hash, retHash, "the returned hash should be accurate")
	if assert.NoError(t, err) {
		assert.Equal(t, "", buffer.String())
	}

	buffer.Reset()
	writer.knownUsers = createKnownUserMap(0)
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
		knownUsers: createKnownUserMap(2),
		outputLock: sync.Mutex{},
		output:     buffer,
	}
	writer.knownUsers["hash1"] = nil
	writer.knownUsers["hash2"] = nil
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
	}

	for _, entry := range data {
		buffer.Reset()
		if assert.NoError(t, writer.WriteEventUserHash(entry.Hash, entry.Location, entry.EventType)) {
			assert.Equal(t, fmt.Sprintf("%s\t%d\n", entry.Result, time.Now().Unix()), buffer.String())
		}
	}

	assert.Equal(t, loc2, writer.knownUsers["hash1"])
	assert.Equal(t, (*Location)(nil), writer.knownUsers["hash2"])

	assert.Error(t, writer.WriteEventUserHash("unknown_hash", loc1, LOGIN), "attempts to write unknown hashes directly should fail")

	ew := newErrorWriter()
	writer.output = &ew
	assert.Error(t, writer.WriteEventUserHash("hash1", loc1, LOGIN), "journal writer errors should be propagated")
}

func TestWriter_WriteEventUser(t *testing.T) {
	t.Parallel()
	loc1 := &Location{Name: "Mosbach", Code: "MOS"}
	loc2 := &Location{Name: "Teststadt", Code: "TST"}
	buffer := bytes.Buffer{}
	writer := Writer{
		knownUsers: createKnownUserMap(2),
		outputLock: sync.Mutex{},
		output:     &buffer,
	}
	user1 := User{Name: "Tester", Address: "TAddr"}
	hash1 := util.Base64Encode(user1.Hash())
	user2 := User{Name: "", Address: ""}
	hash2 := util.Base64Encode(user2.Hash())

	if assert.NoError(t, writer.WriteEventUser(&user1, loc1, LOGIN)) {
		assert.Equal(
			t, fmt.Sprintf("*%s\t%s\n+%s\tMOS\t%d\n", user1.Name, user1.Address, hash1, time.Now().Unix()),
			buffer.String(),
		)
	}

	buffer.Reset()
	if assert.NoError(t, writer.WriteEventUser(&user1, loc1, LOGOUT)) {
		assert.Equal(
			t, fmt.Sprintf("-%s\tMOS\t%d\n", hash1, time.Now().Unix()),
			buffer.String(),
		)
	}

	buffer.Reset()
	writer.knownUsers[hash2] = nil
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

	delete(writer.knownUsers, hash1)
	assert.Error(t, writer.WriteEventUser(&user1, loc1, LOGOUT))
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
