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

	_ = commandGroup.AddSubcommand(argp.CreateSubcommand("help", "Prints this help text"))

	showPersonCmd := commandGroup.AddSubcommand(argp.CreateSubcommand("show-person", "Show the person with the given name"))
	showPersonJournal := showPersonCmd.PositionalString(argp.FlagBuildArgs{
		Names: []string{"journal"},
		Usage: "the journal input path",
	}, "")
	showPersonName := showPersonCmd.PositionalString(argp.FlagBuildArgs{
		Names: []string{"name"},
		Usage: "The name of the person",
	}, "")

	exportCmd := commandGroup.AddSubcommand(argp.CreateSubcommand("export", "Export the journal to CSV"))
	exportJournal := exportCmd.PositionalString(argp.FlagBuildArgs{
		Names: []string{"journal"},
		Usage: "The journal input path",
	}, "")
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

	switch subcommand.Name {
	case "help":
		commandGroup.PrintUsage("")
	case "show-person":
		j := readJournal(*showPersonJournal)
		user := (*journal.User)(nil)
		for iterUser := range j.GetUsers() {
			if strings.Contains(strings.ToLower(iterUser.Name), strings.ToLower(*showPersonName)) {
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
	case "export":
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
		fmt.Printf("Failed to read readJournal (\"%s\"): %v\n", path, err)
		os.Exit(100)
	}
	return &readJournal
}
