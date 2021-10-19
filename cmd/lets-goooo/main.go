package main

import (
	"fmt"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/token"
)

func main() {
	println("Let's goooo!")
	//RunWebservers()
	a, b := token.GetQrCode("https://localhost:4443", "MOS")
	fmt.Printf("%v %v", a, b)
}
