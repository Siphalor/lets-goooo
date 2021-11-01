package cmd

import (
	"fmt"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/journal"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/util"
	"os"
	"time"
)

func ViewContacts(
	journalPath string, locationsPath string, name string, address string, csv bool,
	csvHeaders bool, outputPath string, outputPerms uint) {

	forceReadLocations(locationsPath)
	j := forceReadJournal(journalPath)
	user := findUser(j, name, address)

	writer := openOutput(outputPath, outputPerms)
	defer func() {
		err := writer.Close()
		if err != nil {
			println("Failed to close output")
		}
	}()

	err := error(nil)
	if csv {
		if csvHeaders {
			err = util.WriteString(writer, "Duration in seconds,Location,Contact Name,Contact Address\n")
		}
	} else { // Print helper message with name and address of person
		err = util.WriteString(writer, fmt.Sprintf("Showing contacs for user %s (%s):\n", user.Name, user.Address))
	}
	if err != nil {
		fmt.Printf("Failed to write to output: %v", err)
		os.Exit(500)
	}

	userLogin := (*journal.Event)(nil)         // The last read user login event
	lastLocHeading := (*journal.Location)(nil) // The last written location heading, so locational contacts are grouped together

	// Map of locations and their current users with their login events
	allUserLocs := make(map[*journal.Location]map[*journal.User]*journal.Event, len(journal.Locations))
	// Initialize that map with the known locations
	for _, loc := range journal.Locations {
		allUserLocs[loc] = make(map[*journal.User]*journal.Event, 50)
	}

	// Private function for printing contact information
	printContact := func(otherUser *journal.User, login *journal.Event, logout *journal.Event) {
		// Write location headers only when not in CSV mode and on location changes
		if !csv && lastLocHeading == nil {
			err := util.WriteString(writer, login.Location.Name+":\n")
			if err != nil {
				fmt.Printf("Failed to write to output: %v", err)
				os.Exit(500)
			}
			lastLocHeading = login.Location
		}
		// Calculate the duration between login and logout
		duration := time.Unix(logout.Timestamp, 0).Sub(time.Unix(login.Timestamp, 0))
		secs := int(duration.Seconds())

		err := error(nil)
		if csv {
			err = util.WriteString(writer, fmt.Sprintf("%d,%s,\"%s\",\"%s\"\n", secs, login.Location.Name, otherUser.Name, otherUser.Address))
		} else {
			err = util.WriteString(writer, fmt.Sprintf(
				"  %2dh %2dm %2ds - %s - %s\n",
				secs/3600, secs/60%60, secs%60,
				otherUser.Name, otherUser.Address,
			))
		}
		if err != nil {
			fmt.Printf("Failed to write to output: %v", err)
			os.Exit(500)
		}
	}

	events := j.GetEvents()
	for i, event := range events {
		// If an event concerning the selected user is encountered
		if event.User == user {
			switch event.EventType {
			case journal.LOGIN: // on login just set the login event
				userLogin = &events[i]

			case journal.LOGOUT: // on logout check all other persons that are currently checked in
				for otherUser, otherLogin := range allUserLocs[userLogin.Location] {
					printContact(
						otherUser,
						getEarlierEvent(userLogin, otherLogin),
						&event,
					)
				}
				userLogin = nil
			}

		} else { // If the event is about a different user
			switch event.EventType {
			case journal.LOGIN: // store the login event
				allUserLocs[event.Location][event.User] = &events[i]

			case journal.LOGOUT: // check if the user is at the same location as the selected user, then print that contact
				if userLogin != nil && event.Location == userLogin.Location {
					login := allUserLocs[event.Location][event.User]
					printContact(
						event.User,
						getEarlierEvent(login, userLogin),
						&event,
					)
				}

				// remove login event (check out)
				delete(allUserLocs[event.Location], event.User)
			}
		}
	}
}

// getEarlierEvent returns the event that happened earlier from the given arguments
func getEarlierEvent(evt1 *journal.Event, evt2 *journal.Event) *journal.Event {
	if evt1.Timestamp < evt2.Timestamp {
		return evt1
	}
	return evt2
}
