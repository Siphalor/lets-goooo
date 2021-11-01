package cmd

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func ExampleShowPerson() {
	tz := time.Local
	time.Local = time.UTC
	err := ShowPerson("testdata/journal.txt", "testdata/locations.xml", "Tester", "")
	if err != nil {
		fmt.Printf("Error: %v", err)
	}
	time.Local = tz

	// Output:
	// Teststadt:
	//      Login:  3:20:00
	//     Logout:  3:36:40
	// Hauptstadt:
	//      Login:  4:10:00
	//     Logout: 10:00:00
}

func TestShowPerson(t *testing.T) {
	assert.Error(t, ShowPerson("testdata/missingno", "testdata/locations.xml", "Tester", ""))
	assert.Error(t, ShowPerson("testdata/journal.txt", "testdata/missingno", "Tester", ""))
	assert.Error(t, ShowPerson("testdata/journal.txt", "testdata/locations.xml", "Muad'Dib", ""))
	assert.Error(t, ShowPerson("testdata/journal.txt", "testdata/locations.xml", "", ""))
	assert.Error(t, ShowPerson("testdata/journal.txt", "testdata/locations.xml", "", "???"))
}
