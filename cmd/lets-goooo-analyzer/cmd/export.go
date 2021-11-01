package cmd

import (
	"fmt"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/journal"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/util"
	"strings"
)

func Export(journalPath string, locationsPath string, csvHeaders bool, outputPath string, outputPerms uint, locationFilterName string) *Error {
	err := readLocations(locationsPath)
	if err != nil {
		return err
	}
	var locationFilter *journal.Location = nil
	if locationFilterName != "" {
		location, exists := journal.Locations[locationFilterName]
		if exists {
			locationFilter = location
		} else {
			for _, loc := range journal.Locations {
				if strings.ToLower(loc.Name) == strings.ToLower(locationFilterName) {
					locationFilter = loc
					break
				}
			}

			if locationFilter == nil {
				return NewError(404, fmt.Sprintf("failed to resolve location \"%s\"\n", locationFilterName), nil)
			}
		}
	}

	j, err := readJournal(journalPath)
	if err != nil {
		return err
	}
	if outputPath == "" { // set a default output file name
		outputPath = journalPath + "-export.csv"
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

	if csvHeaders { // Print the CSV headers, if applicable
		err := util.WriteString(writer, "Event type,Location,Timestamp,Name,Address\n")
		if err != nil {
			return NewError(500, "failed to write to output", err)
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
	return nil
}
