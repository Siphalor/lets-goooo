package main

import (
	"fmt"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/argp"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/journal"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/util"
	"os"
	"strings"
	"time"
)

func main() {
	// Subcommand structure based on https://gobyexample.com/command-line-subcommands
	commandGroup := argp.CreateSubcommandGroup()

	helpCmd := commandGroup.AddSubcommand(argp.CreateSubcommand("help", "Prints this help text"))

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

	showPersonCmd := commandGroup.AddSubcommand(argp.CreateSubcommand("show-person", "Show the person with the given name"))
	showPersonJournal := showPersonCmd.PositionalString(journalProtoArg, "")
	showPersonName := showPersonCmd.String(personNameProtoArg, "")
	showPersonAddress := showPersonCmd.String(personAddressProtoArg, "")

	viewContactsCmd := commandGroup.AddSubcommand(argp.CreateSubcommand("view-contacts", "Creates a personal contact list with a journal"))
	viewContactsJournal := viewContactsCmd.PositionalString(journalProtoArg, "")
	viewContactsName := viewContactsCmd.String(personNameProtoArg, "")
	viewContactsAddress := viewContactsCmd.String(personAddressProtoArg, "")

	exportCmd := commandGroup.AddSubcommand(argp.CreateSubcommand("export", "Export the journal to CSV"))
	exportJournal := exportCmd.PositionalString(journalProtoArg, "")
	exportOutputDefault := "<journal>-export.csv"
	exportOutput := exportCmd.String(argp.FlagBuildArgs{
		Names:       []string{"output-file", "output", "o"},
		Usage:       "The CSV output file",
		DefaultText: &exportOutputDefault,
	}, "")
	exportOutputPermsDefault := "0660"
	exportOutputPerms := exportCmd.Uint(argp.FlagBuildArgs{
		Names:       []string{"output-file-perms", "output-perms"},
		Usage:       "The permission mask for the output file",
		DefaultText: &exportOutputPermsDefault,
	}, 0660)

	subcommand, err := commandGroup.ParseSubcommand(os.Args)
	if err != nil {
		return
	}
	if subcommand == nil {
		return
	}

	switch subcommand {
	case helpCmd:
		commandGroup.PrintUsage("")
		os.Exit(0)
	case showPersonCmd:
		j := readJournal(*showPersonJournal)
		user := findUser(j, *showPersonName, *showPersonAddress)

		lastLoc := (*journal.Location)(nil)
		for _, event := range j.GetEvents() {
			if event.User == user {
				if event.Location != lastLoc {
					fmt.Printf("%s:\n", event.Location.Name)
				}
				eventTime := time.Unix(event.Timestamp, 0).In(time.Local)
				fmt.Printf("%10s: %2d:%02d:%02d\n", event.EventType.Name(), eventTime.Hour(), eventTime.Minute(), eventTime.Second())
				lastLoc = event.Location
			}
		}
	case viewContactsCmd:
		j := readJournal(*viewContactsJournal)
		user := findUser(j, *viewContactsName, *viewContactsAddress)
		fmt.Printf("Showing contacts for user %s (%s):\n", user.Name, user.Address)

		userLogin := (*journal.Event)(nil)
		lastLocHeading := (*journal.Location)(nil)

		allUserLocs := make(map[*journal.Location]map[*journal.User]*journal.Event, len(journal.Locations))
		for _, loc := range journal.Locations {
			allUserLocs[loc] = make(map[*journal.User]*journal.Event, 50)
		}

		printContact := func(otherUser *journal.User, login *journal.Event, logout *journal.Event) {
			if lastLocHeading == nil {
				fmt.Printf("%s:\n", login.Location.Name)
				lastLocHeading = login.Location
			}
			duration := time.Unix(logout.Timestamp, 0).Sub(time.Unix(login.Timestamp, 0))
			secs := int(duration.Seconds())
			fmt.Printf(
				"  %2dh %2dm %2ds - %s - %s\n",
				secs/3600, secs/60%60, secs%60,
				otherUser.Name, otherUser.Address,
			)
		}

		events := j.GetEvents()
		for i, event := range events {
			if event.User == user {
				switch event.EventType {
				case journal.LOGIN:
					userLogin = &events[i]

				case journal.LOGOUT:
					for otherUser, otherLogin := range allUserLocs[userLogin.Location] {
						printContact(
							otherUser,
							getEarlierEvent(userLogin, otherLogin),
							&event,
						)
					}
					userLogin = nil
				}

			} else {
				switch event.EventType {
				case journal.LOGIN:
					allUserLocs[event.Location][event.User] = &events[i]

				case journal.LOGOUT:
					if userLogin != nil && event.Location == userLogin.Location {
						login := allUserLocs[event.Location][event.User]
						printContact(
							event.User,
							getEarlierEvent(login, userLogin),
							&event,
						)
					}

					delete(allUserLocs[event.Location], event.User)
				}
			}
		}
	case exportCmd:
		journalPath := *exportJournal
		j := readJournal(journalPath)
		if *exportOutput == "" {
			*exportOutput = journalPath + "-export.csv"
		}
		file, err := os.OpenFile(*exportOutput, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.FileMode(*exportOutputPerms))
		if err != nil {
			fmt.Printf("Failed to open output file: %v\n", err)
			os.Exit(101)
		}
		for _, event := range j.GetEvents() {
			err := util.WriteString(file, fmt.Sprintf(
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
	default:
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

func findUser(j *journal.Journal, nameFilter string, addressFilter string) *journal.User {
	filters := 0
	if nameFilter != "" {
		nameFilter = strings.ToLower(nameFilter)
		filters++
	}
	if addressFilter != "" {
		addressFilter = strings.ToLower(addressFilter)
		filters++
	}
	if filters <= 0 {
		fmt.Printf("Either a filter by name or by address must be specified")
		os.Exit(1)
	}

	user := (*journal.User)(nil)
	for iterUser := range j.GetUsers() {
		matching := 0
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
		if matching >= filters {
			user = iterUser
			break
		}
	}
	if user == nil {
		println("Could not find such a user, known users are:")
		for iterUser := range j.GetUsers() {
			fmt.Printf("\t%s\n", iterUser.Name)
		}
		os.Exit(101)
	}
	return user
}

func getEarlierEvent(evt1 *journal.Event, evt2 *journal.Event) *journal.Event {
	if evt1.Timestamp < evt2.Timestamp {
		return evt1
	}
	return evt2
}
