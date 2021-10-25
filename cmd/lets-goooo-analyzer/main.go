package main

import (
	"fmt"
	"io"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/argp"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/journal"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/util"
	"os"
	"strings"
	"time"
)

func main() {
	commandGroup := argp.CreateSubcommandGroup() // The main subcommand group

	// PROTOTYPES for arguments that are used multiple times
	journalProtoArg := argp.FlagBuildArgs{
		Names: []string{"journal"},
		Usage: "The journal input path",
	}
	personNameProtoArg := argp.FlagBuildArgs{
		Names: []string{"name", "n"},
		Usage: "Find the person by their name",
	}
	personAddressProtoArg := argp.FlagBuildArgs{
		Names: []string{"address", "a"},
		Usage: "Find the person by their address",
	}
	// special override because the output is generated automatically by default
	outputFileProtoArgDefault := "<journal>-export.csv"
	outputFileProtoArg := argp.FlagBuildArgs{
		Names:       []string{"output-file", "output", "o"},
		Usage:       "The CSV output file",
		DefaultText: &outputFileProtoArgDefault,
	}
	// special override to output an uint as octagonal file mode
	outputFilePermsProtoArgDefault := "0660"
	outputFilePermsProtoArg := argp.FlagBuildArgs{
		Names:       []string{"output-file-perms", "output-perms"},
		Usage:       "The permission mask for the output file",
		DefaultText: &outputFilePermsProtoArgDefault,
	}
	csvHeaderProtoArg := argp.FlagBuildArgs{
		Names: []string{"csv-headers", "csv-header-row"},
		Usage: "Whether the CSV file will be prefixed with a header line",
	}

	// HELP command
	helpCmd := commandGroup.AddSubcommand(argp.CreateSubcommand("help", "Prints this help text"))

	// SHOW-PERSON command
	showPersonCmd := commandGroup.AddSubcommand(argp.CreateSubcommand("show-person", "Show the person with the given name"))
	showPersonJournal := showPersonCmd.PositionalString(journalProtoArg, "")
	showPersonName := showPersonCmd.String(personNameProtoArg, "")
	showPersonAddress := showPersonCmd.String(personAddressProtoArg, "")

	// VIEW-CONTACTS command
	viewContactsCmd := commandGroup.AddSubcommand(argp.CreateSubcommand("view-contacts", "Creates a personal contact list with a journal"))
	viewContactsJournal := viewContactsCmd.PositionalString(journalProtoArg, "")
	viewContactsName := viewContactsCmd.String(personNameProtoArg, "")
	viewContactsAddress := viewContactsCmd.String(personAddressProtoArg, "")
	viewContactsCSV := viewContactsCmd.Bool(argp.FlagBuildArgs{
		Names: []string{"csv"},
		Usage: "Output as CSV data, opposed to a human readable format",
	}, false)
	viewContactsCSVHeaders := viewContactsCmd.Bool(csvHeaderProtoArg, false)
	viewContactsOutput := viewContactsCmd.String(outputFileProtoArg, "")
	viewContactsOutputPerms := viewContactsCmd.Uint(outputFilePermsProtoArg, 0660)

	// EXPORT command
	exportCmd := commandGroup.AddSubcommand(argp.CreateSubcommand("export", "Export the journal to CSV"))
	exportJournal := exportCmd.PositionalString(journalProtoArg, "")
	exportCSVHeaders := exportCmd.Bool(csvHeaderProtoArg, false)
	exportOutput := exportCmd.String(outputFileProtoArg, "")
	exportOutputPerms := exportCmd.Uint(outputFilePermsProtoArg, 0660)
	exportLocation := exportCmd.String(argp.FlagBuildArgs{
		Names: []string{"location", "loc", "l"},
		Usage: "Filter the events by a location, given either as code (three letters) or by the full name",
	}, "")

	// Parse the system arguments
	subcommand, err := commandGroup.ParseSubcommand(os.Args[1:])
	if err != nil { // Errors are already printed, no further error handling required
		return
	}
	if subcommand == nil { // The code enters this if a help flag is specified
		return
	}

	switch subcommand {
	case helpCmd:
		commandGroup.PrintUsage("")
		os.Exit(0)

	case showPersonCmd:
		j := readJournal(*showPersonJournal)
		user := findUser(j, *showPersonName, *showPersonAddress)

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

	case viewContactsCmd:
		j := readJournal(*viewContactsJournal)
		user := findUser(j, *viewContactsName, *viewContactsAddress)

		writer := openOutput(*viewContactsOutput, *viewContactsOutputPerms)
		defer func() {
			err := writer.Close()
			if err != nil {
				println("Failed to close output")
			}
		}()

		err := error(nil)
		if *viewContactsCSV {
			if *viewContactsCSVHeaders {
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
			if !*viewContactsCSV && lastLocHeading == nil {
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
			if *viewContactsCSV {
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

	case exportCmd:
		var locationFilter *journal.Location = nil
		if *exportLocation != "" {
			location, exists := journal.Locations[*exportLocation]
			if exists {
				locationFilter = location
			} else {
				for _, loc := range journal.Locations {
					if strings.ToLower(loc.Name) == strings.ToLower(*exportLocation) {
						locationFilter = loc
						break
					}
				}

				if locationFilter == nil {
					fmt.Printf("Failed to resolve location \"%s\"\n", *exportLocation)
					os.Exit(404)
				}
			}
		}

		journalPath := *exportJournal
		j := readJournal(journalPath)
		if *exportOutput == "" { // set a default output file name
			*exportOutput = journalPath + "-export.csv"
		}
		writer := openOutput(*exportOutput, *exportOutputPerms)
		defer func() {
			err := writer.Close()
			if err != nil {
				println("Failed to close output")
			}
		}()

		if *exportCSVHeaders { // Print the CSV headers, if applicable
			err := util.WriteString(writer, "Event type,Location,Timestamp,Name,Address\n")
			if err != nil {
				fmt.Printf("Failed to write to output: %v", err)
				os.Exit(500)
			}
		}
		for _, event := range j.GetEvents() {
			if locationFilter != nil {
				if event.Location != locationFilter {
					continue
				}
			}
			err := util.WriteString(writer, fmt.Sprintf(
				"%s,%s,%d,%s,%s\n",
				event.EventType.Name(),
				event.Location.Name,
				event.Timestamp,
				event.User.Name,
				event.User.Address,
			))
			if err != nil {
				fmt.Printf("Failed to write event to output: %v\n", err)
			}
		}

	default: // should™ be unreachable
		println("Invalid subcommand!")
	}
}

func readJournal(path string) *journal.Journal {
	readJournal, err := journal.ReadJournal(path)
	if err != nil {
		fmt.Printf("Failed to read journal (\"%s\"): %v\n", path, err)
		os.Exit(100)
	}
	return &readJournal
}

// openOutput returns an output stream, either to a new file or to stdout
func openOutput(outputArg string, outputPermsArg uint) io.WriteCloser {
	if outputArg == "" { // If no output file is specified, then use stdout
		return os.Stdout
	}
	// else, try to open a new file at the location with the given properties
	file, err := os.OpenFile(outputArg, os.O_WRONLY|os.O_TRUNC|os.O_APPEND|os.O_CREATE, os.FileMode(outputPermsArg))
	if err != nil {
		fmt.Printf("Failed to open output file %s: %v", outputArg, err)
	}
	return file
}

// findUser tries to find a user with the given filters in the journal
func findUser(j *journal.Journal, nameFilter string, addressFilter string) *journal.User {
	filters := 0 // The count of filters required to match on a user
	if nameFilter != "" {
		nameFilter = strings.ToLower(nameFilter)
		filters++
	}
	if addressFilter != "" {
		addressFilter = strings.ToLower(addressFilter)
		filters++
	}
	if filters <= 0 { // no filters set
		println("Either a filter by name or by address must be specified")
		os.Exit(1)
	}

	user := (*journal.User)(nil) // The potentially found user
	for iterUser := range j.GetUsers() {
		matching := 0 // The number of matched filters
		if nameFilter != "" {
			if strings.Contains(strings.ToLower(iterUser.Name), nameFilter) {
				matching++
			}
		}
		if addressFilter != "" {
			if strings.Contains(strings.ToLower(iterUser.Address), addressFilter) {
				matching++
			}
		}
		if matching >= filters { // If the enough filters matched, use the current user
			user = iterUser
			break
		}
	}
	if user == nil { // no user found
		println("Could not find such a user, known users are:")
		for iterUser := range j.GetUsers() {
			fmt.Printf("\t%s\n", iterUser.Name)
		}
		os.Exit(101)
	}
	return user
}

// getEarlierEvent returns the event that happened earlier from the given arguments
func getEarlierEvent(evt1 *journal.Event, evt2 *journal.Event) *journal.Event {
	if evt1.Timestamp < evt2.Timestamp {
		return evt1
	}
	return evt2
}
