package cmd

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"path"
	"testing"
)

func ExampleExport_stdoutHeaders() {
	err := Export("testdata/journal.txt", "testdata/locations.xml", true, "-", 0777, "")
	if err != nil {
		fmt.Printf("Error: %v", err)
	}

	// Output:
	// Event type,Location,Timestamp,Name,Address
	// Login,Teststadt,1634700000,Tester,Teststadt
	// Logout,Teststadt,1634701000,Tester,Teststadt
	// Login,Hauptstadt,1634703000,Tester,Teststadt
	// Login,Hauptstadt,1634710000,Klaus,Musterdorf
	// Logout,Hauptstadt,1634712000,Klaus,Musterdorf
	// Login,Teststadt,1634720000,Klaus,Musterdorf
	// Logout,Hauptstadt,1634724000,Tester,Teststadt
	// Logout,Teststadt,1634726000,Klaus,Musterdorf
}

func ExampleExport_stdoutFilterLong() {
	err := Export("testdata/journal.txt", "testdata/locations.xml", false, "-", 0777, "Teststadt")
	if err != nil {
		fmt.Printf("Error: %v", err)
	}

	// Output:
	// Login,Teststadt,1634700000,Tester,Teststadt
	// Logout,Teststadt,1634701000,Tester,Teststadt
	// Login,Teststadt,1634720000,Klaus,Musterdorf
	// Logout,Teststadt,1634726000,Klaus,Musterdorf
}

func ExampleExport_stdoutFilterShort() {
	err := Export("testdata/journal.txt", "testdata/locations.xml", false, "-", 0777, "HST")
	if err != nil {
		fmt.Printf("Error: %v", err)
	}

	// Output:
	// Login,Hauptstadt,1634703000,Tester,Teststadt
	// Login,Hauptstadt,1634710000,Klaus,Musterdorf
	// Logout,Hauptstadt,1634712000,Klaus,Musterdorf
	// Logout,Hauptstadt,1634724000,Tester,Teststadt
}

func TestExport_fileOutput(t *testing.T) {
	dir := t.TempDir()
	outFile := path.Join(dir, "out.csv")
	err := Export("testdata/journal.txt", "testdata/locations.xml", false, outFile, 0777, "TST")
	if assert.NoError(t, err) {
		if assert.FileExists(t, outFile) {
			content, err := ioutil.ReadFile(outFile)
			if assert.NoError(t, err) {
				assert.Equal(
					t,
					"Login,Teststadt,1634700000,Tester,Teststadt\n"+
						"Logout,Teststadt,1634701000,Tester,Teststadt\n"+
						"Login,Teststadt,1634720000,Klaus,Musterdorf\n"+
						"Logout,Teststadt,1634726000,Klaus,Musterdorf\n",
					string(content),
				)
			}
		}
	}
}

func TestExport_errors(t *testing.T) {
	tempDir := t.TempDir()
	assert.Error(t, Export("testdata/journal.txt", "testdata/missingno", true, "-", 0777, "TST"))
	assert.Error(t, Export("testdata/journal.txt", "testdata/locations.xml", true, "-", 0777, "???"))
	assert.Error(t, Export("testdata/missingno", "testdata/locations.xml", true, "-", 0777, "TST"))
	assert.Error(t, Export("testdata/journal.txt", "testdata/locations.xml", true, tempDir, 07000, "TST"))
}
