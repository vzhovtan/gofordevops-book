package grpcexample

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func RunClient() {
	// Connect to the gRPC server
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := NewPersonServiceClient(conn)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Create a person
	createResp, err := client.CreatePerson(ctx, &CreatePersonRequest{
		Person: &Person{
			Id:    "1",
			Name:  "John Doe",
			Age:   30,
			Email: "john.doe@example.com",
		},
	})
	if err != nil {
		log.Fatalf("CreatePerson failed: %v", err)
	}
	log.Printf("CreatePerson response: %s", createResp.Message)
	log.Printf("Created person: %v", createResp.Person)

	// Get the person
	getResp, err := client.GetPerson(ctx, &GetPersonRequest{
		Id: "1",
	})
	if err != nil {
		log.Fatalf("GetPerson failed: %v", err)
	}
	if getResp.Person != nil {
		log.Printf("Retrieved person: Name=%s, Age=%d, Email=%s",
			getResp.Person.Name,
			getResp.Person.Age,
			getResp.Person.Email)
	} else {
		log.Println("Person not found")
	}
}
