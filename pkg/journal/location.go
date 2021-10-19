package journal

// Locations is the global list of known locations.
var Locations = map[string]*Location{ // TODO: load dynamically
	"MOS": {Name: "Mosbach", Code: "MOS"},
	"MGH": {Name: "Bad Mergentheim", Code: "MGH"},
}

// Location represents a location where users can sign in to.
type Location struct {
	Name string
	Code string
}
