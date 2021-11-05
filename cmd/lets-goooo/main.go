package main

import "fmt"

func main() {
	println("Let's goooo!")
	err := RunWebservers(4443, 443)
	if err != nil {
		fmt.Printf("couldn't start the Webservers: %#v", err)
	}
}
