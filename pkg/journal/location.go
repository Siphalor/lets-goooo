package journal

var Locations = map[string]*Location{ // TODO: load dynamically
	"MOS": {Name: "Mosbach", Code: "MOS"},
	"MGH": {Name: "Bad Mergentheim", Code: "MGH"},
}

type Location struct {
	Name string
	Code string
}
