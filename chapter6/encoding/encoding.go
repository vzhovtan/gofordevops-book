package encoding

import (
	"encoding/json"
	"fmt"
	"os"
)

// Person struct with various field types and JSON tags
type Person struct {
	Name    string  `json:"name"`
	Age     int     `json:"age"`
	Email   string  `json:"email,omitempty"` // omitempty: exclude if empty
	Address Address `json:"address"`
	Private string  `json:"-"` // "-" means this field is never encoded
}

type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	Country string `json:"country"`
	ZipCode string `json:"zip_code,omitempty"`
}

func EncodingJson() {

	// Creating an instance of Person
	person := Person{
		Name:  "Alice White",
		Age:   30,
		Email: "alice.white@example.com",
		Address: Address{
			Street:  "123 Main St",
			City:    "San Francisco",
			Country: "USA",
			ZipCode: "94105",
		},
		Private: "This will not be encoded",
	}

	// Using json.Encoder - Write directly to io.Writer
	fmt.Println("Using json.Encoder (write to stdout):")
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(person); err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println()
}
