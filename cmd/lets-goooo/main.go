package main

import "lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/token"

func main() {
	println("Let's goooo!")
	println(123)
	println(string(token.GetQrCode("https://localhost:4443", "MOS")))
}
