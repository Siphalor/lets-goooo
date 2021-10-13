package journal

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		assert.Equal(t, GetCurrentJournalPath(tempDir), writer.output.Name(), "wrong file output path")
	}
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
	_, err = file.WriteString("+ignored\n-ignored\n")
	for _, user := range users {
		_, err = file.WriteString(fmt.Sprintf("*%s\t%s\n", user.Name, user.Address))
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
	_, err = file.WriteString("*invalid line")
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
		if assert.Equal(t, GetCurrentJournalPath(tempDir), writer.output.Name(), "incorrect journal file path") {
			assert.FileExists(t, writer.output.Name(), "failed to create journal file")
		}
	}

	writer.directory = path.Join(writer.directory, "none")
	if assert.NoError(t, writer.UpdateOutput(), "failed to run update output for non-existing directory") {
		if assert.Equal(t, GetCurrentJournalPath(writer.directory), writer.output.Name(), "incorrect journal file name") {
			assert.FileExists(t, writer.output.Name(), "failed to create journal file")
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
