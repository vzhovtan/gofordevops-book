package grpcexample

import (
	"context"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
)

type server struct {
	UnimplementedPersonServiceServer
	persons map[string]*Person
	mu      sync.RWMutex
}

func newServer() *server {
	return &server{
		persons: make(map[string]*Person),
	}
}

func (s *server) CreatePerson(ctx context.Context, req *CreatePersonRequest) (*CreatePersonResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	person := req.GetPerson()
	s.persons[person.Id] = person

	log.Printf("Created person: %v", person)

	return &CreatePersonResponse{
		Person:  person,
		Message: "Person created successfully",
	}, nil
}

func (s *server) GetPerson(ctx context.Context, req *GetPersonRequest) (*GetPersonResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	person, exists := s.persons[req.Id]
	if !exists {
		return &GetPersonResponse{
			Person: nil,
		}, nil
	}

	log.Printf("Retrieved person: %v", person)

	return &GetPersonResponse{
		Person: person,
	}, nil
}

func RunServer() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	RegisterPersonServiceServer(grpcServer, newServer())

	log.Println("gRPC server listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
