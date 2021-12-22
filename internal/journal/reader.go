package journal

import (
	"bufio"
	"fmt"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/internal/util"
	"log"
	"os"
)

// Journal is a read-only representation of a journal file.
type Journal struct {
	users  map[string]*User
	events []Event
}

// ReadJournal reads in a Journal from a journal file.
func ReadJournal(filepath string) (Journal, error) {
	if isFile, err := util.FileExists(filepath); err != nil || !isFile {
		return Journal{}, fmt.Errorf("\"%s\" is not a valid file (%w)", filepath, err)
	}
	file, err := os.OpenFile(filepath, os.O_RDONLY, 0777)
	if err != nil {
		return Journal{}, fmt.Errorf("failed to open journal file %s: %w", filepath, err)
	}
	journal := Journal{
		users:  make(map[string]*User, 100),
		events: make([]Event, 0, 1000),
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		switch line[0] {
		case '*':
			user, err := ParseUserJournalLine(line[1:])
			if err != nil {
				return journal, fmt.Errorf("failed to read journal User line \"%s\": %w", line, err)
			}
			journal.users[string(user.Hash())] = &user
		case uint8(LOGIN), uint8(LOGOUT):
			entry, err := ParseEventJournalEntry(EventType(line[0]), line[1:], &journal.users)
			if err != nil {
				log.Printf("Failed to parse journal line \"%s\": %#v", line, err)
			}
			journal.events = append(journal.events, entry)
		}
	}
	return journal, nil
}

// GetUsers provides a way to iterate over all known users.
func (journal *Journal) GetUsers() <-chan *User {
	out := make(chan *User)

	go func() {
		for _, user := range journal.users {
			out <- user
		}
		close(out)
	}()
	return out
}

// GetEvents yields all events in the journal.
func (journal *Journal) GetEvents() []Event {
	return journal.events
}
