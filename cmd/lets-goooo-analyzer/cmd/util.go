package cmd

import (
	"fmt"
	"io"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/journal"
	"os"
	"strings"
)

// forceReadJournal reads the journal at the given path or fails hard
func forceReadJournal(path string) *journal.Journal {
	readJournal, err := journal.ReadJournal(path)
	if err != nil {
		fmt.Printf("Failed to read journal (\"%s\"): %v\n", path, err)
		os.Exit(100)
	}
	return &readJournal
}

// forceReadLocations reads the locations file at the given path or fails hard
func forceReadLocations(arg string) {
	if arg != "" {
		if err := journal.ReadLocations(arg); err != nil {
			fmt.Printf("Failed to read locations from file \"%s\": %v", arg, err)
			os.Exit(103)
		}
	}
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
		os.Exit(104)
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
