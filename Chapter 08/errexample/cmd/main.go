package main

import (
	"errexample"
	"fmt"
	"log"
)

func main() {
	//errexample.FileOpenRead("nonexisting.txt")
	content, err := errexample.FileOpenRead2("nonexisting.txt")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(content)
}
