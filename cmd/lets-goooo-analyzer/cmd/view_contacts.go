// Part of the Let's Goooo project
// Copyright 2021; matriculation numbers: 1103207, 3106445, 4485500
// Let's goooo get this over together

package cmd

import (
	"fmt"
	"io"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/internal/journal"
	"time"
)

func ViewContacts(
	journalPath string, locationsPath string, name string, address string, csv bool,
	csvHeaders bool, outputPath string, outputPerms uint) error {

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

	writer, err := openOutput(outputPath, outputPerms)
	if err != nil {
		return err
	}
	defer func() {
		err := writer.Close()
		if err != nil {
			println("Failed to close output")
		}
	}()

	if csv {
		if csvHeaders {
			err = writeString(writer, "Duration in seconds,Location,Contact Name,Contact Address\n")
			if err != nil {
				return err
			}
		}
	} else { // Print helper message with name and address of person
		err = writeString(writer, fmt.Sprintf("Showing contacts for user %s (%s):\n", user.Name, user.Address))
		if err != nil {
			return err
		}
	}

	userLogin := (*journal.Event)(nil)         // The last read user login event
	lastLocHeading := (*journal.Location)(nil) // The last written location heading, so locational contacts are grouped together

	// Map of locations and their current users with their login events
	allUserLocs := make(map[*journal.Location]map[*journal.User]*journal.Event, len(journal.Locations))
	// Initialize that map with the known locations
	for _, loc := range journal.Locations {
		allUserLocs[loc] = make(map[*journal.User]*journal.Event, 50)
	}

	events := j.GetEvents()
	for i, event := range events {
		// If an event concerning the selected user is encountered
		if event.User == user {
			switch event.EventType {
			case journal.LOGIN: // on login just set the login event
				userLogin = &events[i]

			case journal.LOGOUT: // on logout check all other persons that are currently checked in
				if userLogin == nil { // handle unexpected logout
					continue
				}
				for otherUser, otherLogin := range allUserLocs[userLogin.Location] {
					err = printContact(
						writer,
						otherUser, getLaterEvent(userLogin, otherLogin), &event,
						csv, &lastLocHeading,
					)
					if err != nil {
						return err
					}
				}
				userLogin = nil
			}

		} else { // If the event is about a different user
			switch event.EventType {
			case journal.LOGIN: // store the login event
				allUserLocs[event.Location][event.User] = &events[i]

			case journal.LOGOUT: // check if the user is at the same location as the selected user, then print that contact
				if userLogin != nil && event.Location == userLogin.Location {
					login, exists := allUserLocs[event.Location][event.User]
					if !exists { // handle unexpected logout
						continue
					}
					err = printContact(
						writer,
						event.User, getLaterEvent(login, userLogin), &event,
						csv, &lastLocHeading,
					)
					if err != nil {
						return err
					}
				}

				// remove login event (check out)
				delete(allUserLocs[event.Location], event.User)
			}
		}
	}

	return nil
}

// getLaterEvent returns the event that happened earlier from the given arguments
func getLaterEvent(evt1 *journal.Event, evt2 *journal.Event) *journal.Event {
	if evt1.Timestamp > evt2.Timestamp {
		return evt1
	}
	return evt2
}

func printContact(
	writer io.Writer, otherUser *journal.User, login *journal.Event, logout *journal.Event,
	csv bool, lastLocHeading **journal.Location,
) error {
	// Write location headers only when not in CSV mode and on location changes
	if !csv && (*lastLocHeading == nil || *lastLocHeading != login.Location) {
		err := writeString(writer, login.Location.Name+":\n")
		if err != nil {
			return err
		}
		*lastLocHeading = login.Location
	}
	// Calculate the duration between login and logout
	duration := time.Unix(logout.Timestamp, 0).Sub(time.Unix(login.Timestamp, 0))
	secs := int(duration.Seconds())

	if csv {
		err := writeString(writer, fmt.Sprintf("%d,%s,\"%s\",\"%s\"\n", secs, login.Location.Name, otherUser.Name, otherUser.Address))
		if err != nil {
			return err
		}
	} else {
		err := writeString(writer, fmt.Sprintf(
			"  %2dh %2dm %2ds - %s - %s\n",
			secs/3600, secs/60%60, secs%60,
			otherUser.Name, otherUser.Address,
		))
		if err != nil {
			return err
		}
	}

	return nil
}
