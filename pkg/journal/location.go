package journal

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
)

var Locations map[string]*Location

type Location struct {
	XMLName xml.Name `xml:"location"`
	Name    string   `xml:"name,attr"`
	Code    string   `xml:"code,attr"`
}

type locationsXML struct {
	XMLName   xml.Name   `xml:"locations"`
	Locations []Location `xml:"location"`
}

func ReadLocations(path string) error {
	xmlFile, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}

	defer func() {
		err = xmlFile.Close()
		if err != nil {
			fmt.Printf("error occured during closing of Location XML file: %v", err)
		}
	}()

	byteValue, _ := ioutil.ReadAll(xmlFile)

	l := locationsXML{}
	err = xml.Unmarshal(byteValue, &l)
	if err != nil {
		return fmt.Errorf("error occured during loading of Location XML file: %v", err)
	}
	Locations = map[string]*Location{}

	for _, location := range l.Locations {
		Locations[location.Code] = &location
	}
	return nil
}

/*
<locations>
	<location name="Mosbach" code="MOS"></location>
	<location name="Bad Mergentheim" code="MGH"></location>
</locations>
*/
