package sqlexample

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Person struct {
	ID   int
	Name string
	Age  int
}

func SQLCRUD() {
	// Open or create database
	db, err := sql.Open("sqlite3", "test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create table
	createTable := `
	CREATE TABLE IF NOT EXISTS people (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		age INTEGER NOT NULL
	);`

	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Table created successfully")

	// Insert data
	insertStmt := "INSERT INTO people (name, age) VALUES (?, ?)"
	result, err := db.Exec(insertStmt, "Alice", 30)
	if err != nil {
		log.Fatal(err)
	}
	id, _ := result.LastInsertId()
	fmt.Printf("Inserted record with ID: %d\n", id)

	db.Exec(insertStmt, "Bob", 25)
	db.Exec(insertStmt, "Charlie", 35)

	// Query data
	rows, err := db.Query("SELECT id, name, age FROM people")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("\nAll people:")
	for rows.Next() {
		var p Person
		err := rows.Scan(&p.ID, &p.Name, &p.Age)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ID: %d, Name: %s, Age: %d\n", p.ID, p.Name, p.Age)
	}

	// Query single row
	var name string
	var age int
	err = db.QueryRow("SELECT name, age FROM people WHERE name = ?", "Alice").Scan(&name, &age)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nFound: %s is %d years old\n", name, age)

	// Update data
	_, err = db.Exec("UPDATE people SET age = ? WHERE name = ?", 31, "Alice")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Updated Alice's age to 31")

	// Delete data
	_, err = db.Exec("DELETE FROM people WHERE name = ?", "Bob")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Deleted Bob from database")
}
