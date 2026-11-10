package marshalling

import (
	"encoding/json"
	"fmt"
	"log"
)

type Person struct {
	Name string
	Age  int
	City string
}

func BasicMarshalling() {
	person := Person{
		Name: "Alice",
		Age:  30,
		City: "New York",
	}

	jsonData, err := json.Marshal(person)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("--- 0. Regular Marshaling ---")
	fmt.Println(string(jsonData))
}
