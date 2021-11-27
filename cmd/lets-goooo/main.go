package main

import (
	"fmt"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/journal"
)

func main() {
	journal.ReadLocations("locations.xml")
	dataJournal, _ = journal.NewWriter("something")
	println("Let's goooo!")
	err := RunWebservers(4443, 443)
	if err != nil {
		fmt.Printf("couldn't start the Webservers: %#v", err)
	}
}
