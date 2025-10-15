package tags

import (
	"encoding/json"
	"fmt"
	"log"
)

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
	City string `json:"-"`
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

	fmt.Println(string(jsonData))
}
