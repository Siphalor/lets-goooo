package journal

import (
	"bufio"
	"fmt"
	"io"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/util"
	"log"
	"os"
	"path"
	"sync"
	"time"
)

// Writer is a write-only class to write to journal files.
type Writer struct {
	// knownUsers contains the hashes for the currently known users.
	knownUsers util.StringSet
	// directory is the base directory for the journal files
	directory string
	// outputLock is a mutex for using the output in a thread-safe way.
	// It needs to be locked for mutation as well as writes to the output.
	outputLock sync.Mutex
	// output is the current output stream for the journal.
	// It is usually a file but this should not be relied upon.
	output io.Writer
}

// NewWriter creates a new Writer with the given base directory where journal files will be stored.
// If a file for the current date already exists, it'll recover the data and append to that file.
func NewWriter(directory string) (*Writer, error) {
	writer := Writer{
		knownUsers: util.NewStringSet(100),
		directory:  directory,
	}

	filePath := GetCurrentJournalPath(writer.directory)
	if exists, err := util.FileExists(filePath); exists {
		if err := writer.LoadFrom(filePath); err != nil {
			return nil, fmt.Errorf("failed to parse existing journal data data: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed trying to check for existing journal data: %w", err)
	}

	err := writer.UpdateOutput()
	if err != nil {
		return &writer, fmt.Errorf("failed to create new journal writer: %w", err)
	}
	return &writer, nil
}

// GetCurrentJournalPath determines the current journal output file path for today based on the given journal directory.
func GetCurrentJournalPath(directory string) string {
	return path.Join(directory, util.GetDateFilename(time.Now())+".txt")
}

// LoadFrom extracts already known users from the given journal file.
func (writer *Writer) LoadFrom(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0777)
	if err != nil {
		return fmt.Errorf("failed to open file for reading existing data for journal writer: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Failed to close file %s", filePath)
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		switch line[0] {
		case '*': // line indicating new User
			user, err := ParseUserJournalLine(line[1:])
			if err != nil {
				log.Printf("Failed to parse user line \"%s\"", line[1:])
			}
			writer.knownUsers.Add(string(util.Hash(user)))
		}
	}

	return nil
}

// UpdateOutput updates the output journal file to the current date.
func (writer *Writer) UpdateOutput() error {
	writer.outputLock.Lock()
	defer writer.outputLock.Unlock()
	if closable, ok := writer.output.(io.Closer); ok {
		if err := closable.Close(); err != nil {
			return fmt.Errorf("failed to close journal output: %w", err)
		}
	}
	writer.output = nil
	filePath := GetCurrentJournalPath(writer.directory)
	err := os.MkdirAll(path.Dir(filePath), 0777) // TODO
	if err != nil {
		return fmt.Errorf("failed to create directories for journal: %w", err)
	}
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777) //TODO: variable file perms
	if err != nil {
		return fmt.Errorf("failed to open journal file \"%s\": %w", filePath, err)
	}
	writer.output = file
	return nil
}

// writeLine writes a line to the journal.
// It is thread-safe.
func (writer *Writer) writeLine(line string) error {
	writer.outputLock.Lock()
	err := util.WriteString(writer.output, line+"\n")
	writer.outputLock.Unlock()
	if err != nil {
		return fmt.Errorf("failed to write journal line: %w", err)
	}
	return nil
}

// writeUser writes the given User data to the journal.
func (writer *Writer) writeUser(user *User) error {
	if err := writer.writeLine("*" + user.ToJournalLine()); err != nil {
		return fmt.Errorf("failed to write User data: %w", err)
	}
	return nil
}

// WriteUserIfUnknown writes the given User data to the journal if it's not already present.
func (writer *Writer) WriteUserIfUnknown(user *User) (string, error) {
	hash := util.Base64Encode(util.Hash(user))
	if !writer.knownUsers.Contains(hash) {
		if err := writer.writeUser(user); err != nil {
			return hash, fmt.Errorf("failed to write User data if unknown: %w", err)
		}
		return hash, nil
	}
	return hash, nil
}

// WriteEventUserHash writes an event with the given type and User hash.
func (writer *Writer) WriteEventUserHash(userHash string, location *Location, eventType EventType) error {
	err := writer.writeLine(fmt.Sprintf("%s%s\t%s\t%d", eventType.ToString(), userHash, location.Code, time.Now().UTC().Unix()))
	if err != nil {
		return fmt.Errorf("failed to write User event (type: %v): %w", eventType, err)
	}
	return nil
}

// WriteEventUser writes an event with the given event type for the User to the log.
// If the User does not exist yet, a User line is printed first.
func (writer *Writer) WriteEventUser(user *User, location *Location, eventType EventType) error {
	hash, err := writer.WriteUserIfUnknown(user)
	if err != nil {
		return fmt.Errorf("failed to write User login with User data: %w", err)
	}
	if err = writer.WriteEventUserHash(hash, location, eventType); err != nil {
		return fmt.Errorf("failed to write User login with User data: %w", err)
	}
	return nil
}

// TrackJournalRotation takes care of daily updating the journal file.
// This method should be run as its own routine:
func (writer *Writer) TrackJournalRotation() {
	for {
		now := time.Now().In(time.Local)
		nextDay := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.Local)
		time.Sleep(nextDay.Sub(now))
		for {
			err := writer.UpdateOutput()
			if err == nil {
				break
			}
			time.Sleep(30 * time.Second)
			log.Printf("failed to update journal output: %#v", err)
		}
	}
}
