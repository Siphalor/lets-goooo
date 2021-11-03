package journal

import (
	"fmt"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/util"
	"strconv"
	"strings"
)

type User struct {
	Name    string
	Address string
}

// ParseUserJournalLine parses the journal format of User data into a User struct.
func ParseUserJournalLine(line string) (User, error) {
	parts := strings.SplitN(line, "\t", 3)
	if len(parts) != 2 {
		return User{}, fmt.Errorf("user line should contain exactly two fields")
	}

	return User{
		Name:    parts[0],
		Address: parts[1],
	}, nil
}

// ToJournalLine converts the User to the journal format.
func (user *User) ToJournalLine() string {
	return fmt.Sprintf(
		"%s\t%s",
		strings.ReplaceAll(user.Name, "\t", "    "),
		strings.ReplaceAll(user.Address, "\t", "    "))
}

// Hash creates the hash value for the user, e.g. used in journal files.
func (user *User) Hash() []byte {
	return util.HashString(user.ToJournalLine())
}

// EventType define known types of events.
type EventType rune

const (
	LOGIN  EventType = '+'
	LOGOUT           = '-'
)

func (et EventType) ToString() string {
	return string(et)
}

func (et EventType) Name() string {
	switch et {
	case LOGIN:
		return "Login"
	case LOGOUT:
		return "Logout"
	default:
		return "Unknown event type"
	}
}

// Event is the representation of a User related event.
type Event struct {
	EventType EventType
	User      *User
	Location  *Location
	Timestamp int64
}

// ParseEventJournalEntry parses the event data in journal format into an Event.
// The "users" argument is used to look up the user hash in the known users.
func ParseEventJournalEntry(eventType EventType, data string, users *map[string]*User) (Event, error) {
	parts := strings.SplitN(data, "\t", 3)
	if len(parts) < 3 {
		return Event{}, fmt.Errorf("event data does not contain enough fields")
	}
	hash, err := util.Base64Decode(parts[0])
	if err != nil {
		return Event{}, fmt.Errorf("failed to decode user hash")
	}
	user, exists := (*users)[string(hash)]
	if !exists {
		return Event{}, fmt.Errorf("couldn't resolve User hash \"%s\" in event data", parts[0])
	}
	loc, exists := Locations[parts[1]]
	if !exists {
		return Event{}, fmt.Errorf("couldn't resolve loc code \"%s\"", parts[1])
	}
	unixSeconds, err2 := strconv.ParseInt(parts[2], 10, 64)
	if err2 != nil {
		return Event{}, fmt.Errorf("failed to parse event timestamp \"%s\": %w", parts[1], err2)
	}
	return Event{
		EventType: eventType,
		User:      user,
		Location:  loc,
		Timestamp: unixSeconds,
	}, nil
}
