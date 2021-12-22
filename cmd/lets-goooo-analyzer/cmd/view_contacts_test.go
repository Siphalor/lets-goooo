package cmd

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/internal/journal"
	"testing"
)

func ExampleViewContacts_filterA() {
	err := ViewContacts("testdata/journal_contacts.txt", "testdata/locations.xml", "Tester", "", false, false, "-", 0777)
	if err != nil {
		fmt.Printf("Error: %v", err)
	}

	// Output:
	// Showing contacts for user Tester (Teststadt):
	// Teststadt:
	//    0h  0m  1s - Klaus - Musterdorf
	//    0h 16m 40s - Klaus - Musterdorf
	// Hauptstadt:
	//    0h 16m 40s - Klaus - Musterdorf
	//    1h  0m  1s - Klaus - Musterdorf
}

func ExampleViewContacts_filterA_csv() {
	err := ViewContacts("testdata/journal_contacts.txt", "testdata/locations.xml", "", "Teststadt", true, false, "-", 0777)
	if err != nil {
		fmt.Printf("Error: %v", err)
	}

	// Output:
	// 1,Teststadt,"Klaus","Musterdorf"
	// 1000,Teststadt,"Klaus","Musterdorf"
	// 1000,Hauptstadt,"Klaus","Musterdorf"
	// 3601,Hauptstadt,"Klaus","Musterdorf"
}

func ExampleViewContacts_filterB_csv() {
	err := ViewContacts("testdata/journal.txt", "testdata/locations.xml", "Klaus", "", true, true, "-", 0777)
	if err != nil {
		fmt.Printf("Error: %v", err)
	}

	// Output:
	// Duration in seconds,Location,Contact Name,Contact Address
	// 2000,Hauptstadt,"Tester","Teststadt"
}

func TestViewContacts_errors(t *testing.T) {
	tempDir := t.TempDir()
	assert.Error(t, ViewContacts("testdata/missingno", "testdata/locations.xml", "Klaus", "", false, false, "", 0777))
	assert.Error(t, ViewContacts("testdata/journal.txt", "testdata/missingno", "Klaus", "", false, false, "", 0777))
	assert.Error(t, ViewContacts("testdata/journal.txt", "testdata/locations.xml", "Unknown user", "", false, false, "", 0777))
	assert.Error(t, ViewContacts("testdata/journal.txt", "testdata/locations.xml", "", "Unknown address", false, false, "", 0777))
	assert.Error(t, ViewContacts("testdata/journal.txt", "testdata/locations.xml", "Klaus", "Teststadt", false, false, "", 0777))
	assert.Error(t, ViewContacts("testdata/journal.txt", "testdata/locations.xml", "Klaus", "", false, false, tempDir, 0777))
}

func TestGetLaterEvent(t *testing.T) {
	evt1 := journal.Event{
		Timestamp: 300,
	}
	evt2 := journal.Event{
		Timestamp: 400,
	}
	assert.Equal(t, &evt2, getLaterEvent(&evt1, &evt2))
	assert.Equal(t, &evt2, getLaterEvent(&evt2, &evt1))
}
