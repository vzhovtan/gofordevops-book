package protobuf

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"log"
	"protobuf/example/pb"
)

func PersonMarshalUnmarshal() {
	// Create a Person
	person := &pb.Person{
		Name:  "Alice",
		Id:    1,
		Email: "alice@example.com",
	}

	// Marshal to bytes
	data, err := proto.Marshal(person)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Serialized: %v\n", data)

	// Unmarshal back to struct
	newPerson := &pb.Person{}
	if err := proto.Unmarshal(data, newPerson); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Deserialized: %v\n", newPerson)
}
