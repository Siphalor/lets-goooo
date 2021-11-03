package main

import (
	"encoding/base64"
	"fmt"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/token"
)

func main() {
	println("Let's goooo!")
	//RunWebservers()
	a, b := token.GetQrCode("https://localhost:4443", "MOS")
	fmt.Printf("%v %v", base64.StdEncoding.EncodeToString(a), b)
}
