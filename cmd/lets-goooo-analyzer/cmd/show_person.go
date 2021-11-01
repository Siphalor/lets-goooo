package cmd

import (
	"fmt"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/journal"
	"time"
)

func ShowPersons(journalPath string, locationsPath string, name string, address string) *Error {
	if err := readLocations(locationsPath); err != nil {
		return err
	}
	j, err := readJournal(journalPath)
	if err != nil {
		return err
	}
	user, err := findUser(j, name, address)
	if err != nil {
		return err
	}

	lastLoc := (*journal.Location)(nil) // The last location so that there can be header lines for each location
	for _, event := range j.GetEvents() {
		if event.User == user {
			if event.Location != lastLoc { // Different location
				fmt.Printf("%s:\n", event.Location.Name)
			}
			eventTime := time.Unix(event.Timestamp, 0).In(time.Local) // Important because of daylight saving time or similar happenings
			fmt.Printf("%10s: %2d:%02d:%02d\n", event.EventType.Name(), eventTime.Hour(), eventTime.Minute(), eventTime.Second())
			lastLoc = event.Location
		}
	}

	return nil
}
