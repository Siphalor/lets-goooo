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

func ReadLocations(path string) (eror error) {
	xmlFile, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error occured opening the XML file: %v", err)
	}

	defer func() {
		err = xmlFile.Close()
		if err != nil {
			eror = fmt.Errorf("error occured during closing of Location XML file: %v", err)
		}
	}()

	byteValue, _ := ioutil.ReadAll(xmlFile)

	l := locationsXML{}
	err = xml.Unmarshal(byteValue, &l)
	if err != nil {
		return fmt.Errorf("error occured during loading of Location XML file: %v", err)
	}
	Locations = map[string]*Location{}

	for i, location := range l.Locations {
		Locations[location.Code] = &l.Locations[i]
	}
	return nil
}

/*
<locations>
	<location name="Mosbach" code="MOS"></location>
	<location name="Bad Mergentheim" code="MGH"></location>
</locations>
*/
