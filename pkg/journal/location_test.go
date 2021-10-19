package journal

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/util"
	"os"
	"path"
	"testing"
)

func TestReadLocations(t *testing.T) {

	//Fail - Opening wrong path
	err := ReadLocations("notACorrectPath")
	assert.Error(t, err, "Method did not fail with a wrong path")

	//Fail - Path to a wrong file
	err = ReadLocations("journal.go")
	assert.Error(t, err, "Method did not fail with a wrong file")

	//Extracting correct Information from tmp xml file
	//Creating Tmp Directory for not changing variables
	tempDir, destroy := CreateTempDir(t)
	defer destroy()

	dirPath := path.Join(tempDir, "dir")
	err = os.Mkdir(dirPath, 0777)
	require.NoError(t, err, "internal error: failed to create test directory")

	filepath := path.Join(tempDir, "locations.xml")
	file, _ := os.OpenFile(filepath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0777)

	_ = util.WriteString(file, fmt.Sprintf("<locations>"))
	_ = util.WriteString(file, fmt.Sprintf("    <location name=\"Mosbach\" code=\"MOS\"/>"))
	_ = util.WriteString(file, fmt.Sprintf("    <location name=\"Bad Mergentheim\" code=\"MGH\"/>"))
	_ = util.WriteString(file, fmt.Sprintf("</locations>"))
	_ = file.Close()

	expectedLocationMOS := Location{Name: "Mosbach", Code: "MOS"}
	expectedLocationMGH := Location{Name: "Bad Mergentheim", Code: "MGH"}

	err = ReadLocations("locations.xml")
	assert.NoError(t, err, "Error with correct path")
	assert.Equal(t, 2, len(Locations))

	assert.Equal(t, expectedLocationMOS.Name, Locations["MOS"].Name)
	assert.Equal(t, expectedLocationMOS.Code, Locations["MOS"].Code)
	assert.Equal(t, expectedLocationMGH.Name, Locations["MGH"].Name)
	assert.Equal(t, expectedLocationMGH.Code, Locations["MGH"].Code)

}
