package cmd

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func ExampleShowPerson() {
	err := ShowPerson("testdata/journal.txt", "testdata/locations.xml", "Tester", "")
	if err != nil {
		fmt.Printf("Error: %v", err)
	}

	// Output:
	// Teststadt:
	//      Login:  5:20:00
	//     Logout:  5:36:40
	// Hauptstadt:
	//      Login:  6:10:00
	//     Logout: 12:00:00
}

func TestShowPerson(t *testing.T) {
	assert.Error(t, ShowPerson("testdata/missingno", "testdata/locations.xml", "Tester", ""))
	assert.Error(t, ShowPerson("testdata/journal.txt", "testdata/missingno", "Tester", ""))
	assert.Error(t, ShowPerson("testdata/journal.txt", "testdata/locations.xml", "Muad'Dib", ""))
	assert.Error(t, ShowPerson("testdata/journal.txt", "testdata/locations.xml", "", ""))
	assert.Error(t, ShowPerson("testdata/journal.txt", "testdata/locations.xml", "", "???"))
}
