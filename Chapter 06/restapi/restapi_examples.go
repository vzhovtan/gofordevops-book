package restapi

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

// User struct to hold user data.
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Global slice to store user data.
var users = []User{
	{ID: 1, Name: "Alice White", Email: "alice.white@example.com"},
	{ID: 2, Name: "Jane Smith", Email: "jane.smith@example.com"},
}

// Get all users.
func GetUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// Get a single user by ID.
func GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r) // Use mux.Vars to get route parameters
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest) // Return 400 for bad request
		return
	}

	for _, user := range users {
		if user.ID == id {
			json.NewEncoder(w).Encode(user)
			return
		}
	}

	http.Error(w, "User not found", http.StatusNotFound) // Return 404 if not found
}

// Create a new user.
func CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var newUser User
	err := json.NewDecoder(r.Body).Decode(&newUser) // Use json.NewDecoder
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if newUser.Name == "" || newUser.Email == "" {
		http.Error(w, "Name and email are required", http.StatusBadRequest)
		return
	}

	// Assign a new ID.  In a real application, this would come from a database.
	newUser.ID = getNextUserID() //Call the helper function
	users = append(users, newUser)
	w.WriteHeader(http.StatusCreated) // Return 201 Created status
	json.NewEncoder(w).Encode(newUser)
}

// Helper function to get the next available user ID
func getNextUserID() int {
	maxID := 0
	for _, user := range users {
		if user.ID > maxID {
			maxID = user.ID
		}
	}
	return maxID + 1
}

// Update an existing user.
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var updatedUser User
	err = json.NewDecoder(r.Body).Decode(&updatedUser)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if updatedUser.Name == "" || updatedUser.Email == "" {
		http.Error(w, "Name and email are required", http.StatusBadRequest)
		return
	}

	for i, user := range users {
		if user.ID == id {
			// Update the user data
			users[i].Name = updatedUser.Name
			users[i].Email = updatedUser.Email
			json.NewEncoder(w).Encode(users[i])
			return
		}
	}

	http.Error(w, "User not found", http.StatusNotFound)
}

// Delete a user.
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	for i, user := range users {
		if user.ID == id {
			// Delete the user from the slice
			users = append(users[:i], users[i+1:]...) // Efficient deletion
			w.WriteHeader(http.StatusNoContent)       // 204 No Content
			return
		}
	}

	http.Error(w, "User not found", http.StatusNotFound)
}

func RestApiServer() {
	// Create a new router using gorilla/mux.
	r := mux.NewRouter()

	// Define API endpoints.  Use constants for the paths.
	const usersPath = "/users"
	const userByIDPath = "/users/{id}"

	// Register the handlers.  Use method chaining for cleaner syntax.
	r.HandleFunc(usersPath, GetUsers).Methods(http.MethodGet)
	r.HandleFunc(userByIDPath, GetUser).Methods(http.MethodGet)
	r.HandleFunc(usersPath, CreateUser).Methods(http.MethodPost)
	r.HandleFunc(userByIDPath, UpdateUser).Methods(http.MethodPut)
	r.HandleFunc(userByIDPath, DeleteUser).Methods(http.MethodDelete)

	// Start the server.
	const port = ":9000"
	log.Printf("Server listening on port %s\n", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
