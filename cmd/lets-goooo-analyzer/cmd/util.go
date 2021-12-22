package cmd

import (
	"fmt"
	"io"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/internal/journal"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/internal/util"
	"os"
	"strings"
)

// readJournal reads the journal at the given path
func readJournal(path string) (*journal.Journal, error) {
	readJournal, err := journal.ReadJournal(path)
	if err != nil {
		return nil, NewError(500, fmt.Sprintf("failed to read journal \"%s\"", path), err)
	}
	return &readJournal, nil
}

// readLocations reads the locations file at the given path
func readLocations(arg string) error {
	if arg != "" {
		if err := journal.ReadLocations(arg); err != nil {
			return NewError(500, fmt.Sprintf("failed to read locations from file \"%s\"", arg), err)
		}
	}
	return nil
}

// openOutput returns an output stream, either to a new file or to stdout
func openOutput(outputArg string, outputPermsArg uint) (io.WriteCloser, error) {
	if outputArg == "" || outputArg == "-" { // If no output file is specified, then use stdout
		return os.Stdout, nil
	}
	// else, try to open a new file at the location with the given properties
	file, err := os.OpenFile(outputArg, os.O_WRONLY|os.O_TRUNC|os.O_APPEND|os.O_CREATE, os.FileMode(outputPermsArg))
	if err != nil {
		return nil, NewError(500, fmt.Sprintf("failed to open output file %s", outputArg), err)
	}
	return file, nil
}

func writeString(writer io.Writer, text string) error {
	err := util.WriteString(writer, text)
	if err != nil {
		return NewError(500, "failed to write text to output", err)
	}
	return nil
}

// findUser tries to find a user with the given filters in the journal
func findUser(j *journal.Journal, nameFilter string, addressFilter string) (*journal.User, error) {
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
		return nil, NewError(400, "either a filter by name or by address must be specified", nil)
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
		users := ""
		for iterUser := range j.GetUsers() {
			users += iterUser.Name + ", "
		}
		if users != "" {
			users = users[:len(users)-2]
		}
		return nil, NewError(404, "Could not find such a user, known users are: "+users, nil)
	}
	return user, nil
}
