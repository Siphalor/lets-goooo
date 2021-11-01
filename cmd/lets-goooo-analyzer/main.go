package main

import (
	"lehre.mosbach.dhbw.de/lets-goooo/v2/cmd/lets-goooo-analyzer/cmd"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/argp"
	"os"
)

func main() {
	commandGroup := argp.CreateSubcommandGroup() // The main subcommand group

	// PROTOTYPES for arguments that are used multiple times
	locationsProtoArg := argp.FlagBuildArgs{
		Names: []string{"locations", "l"},
		Usage: "A location XML file to load the location data from",
	}
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
	showPersonLocations := showPersonCmd.String(locationsProtoArg, "locations.xml")
	showPersonName := showPersonCmd.String(personNameProtoArg, "")
	showPersonAddress := showPersonCmd.String(personAddressProtoArg, "")

	// VIEW-CONTACTS command
	viewContactsCmd := commandGroup.AddSubcommand(argp.CreateSubcommand("view-contacts", "Creates a personal contact list with a journal"))
	viewContactsJournal := viewContactsCmd.PositionalString(journalProtoArg, "")
	viewContactsLocations := viewContactsCmd.String(locationsProtoArg, "locations.xml")
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
	exportLocations := exportCmd.String(locationsProtoArg, "locations.xml")
	exportCSVHeaders := exportCmd.Bool(csvHeaderProtoArg, false)
	exportOutput := exportCmd.String(outputFileProtoArg, "")
	exportOutputPerms := exportCmd.Uint(outputFilePermsProtoArg, 0660)
	exportLocation := exportCmd.String(argp.FlagBuildArgs{
		Names: []string{"location", "loc"},
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
		handleCmdError(cmd.ShowPersons(*showPersonJournal, *showPersonLocations, *showPersonName, *showPersonAddress))

	case viewContactsCmd:
		handleCmdError(cmd.ViewContacts(
			*viewContactsJournal, *viewContactsLocations, *viewContactsName, *viewContactsAddress,
			*viewContactsCSV, *viewContactsCSVHeaders, *viewContactsOutput, *viewContactsOutputPerms,
		))

	case exportCmd:
		handleCmdError(cmd.Export(
			*exportJournal, *exportLocations, *exportCSVHeaders, *exportOutput, *exportOutputPerms,
			*exportLocation,
		))

	default: // should™ be unreachable
		println("Invalid subcommand!")
	}
}

func handleCmdError(error *cmd.Error) {
	if error != nil {
		print(error.Error())
		os.Exit(error.Code())
	}
}
