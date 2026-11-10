package sqlexample

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

type PersonTransaction struct {
	ID   int
	Name string
	Age  int
}

// InsertPeopleTransaction inserts multiple people atomically using a transaction.
func InsertPeopleTransaction() {
	db, err := sql.Open("sqlite", "test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 1. Begin the transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}

	// Defer a rollback. If the function returns before tx.Commit(),
	// this ensures pending changes are discarded.
	// If tx.Commit() succeeds, calling tx.Rollback() is a safe no-op.
	defer tx.Rollback()

	insertStmt := "INSERT INTO people (name, age) VALUES (?, ?)"

	// 2. Execute statements using the transaction instance (*sql.Tx)
	_, err = tx.Exec(insertStmt, "Dave", 40)
	if err != nil {
		// The deferred tx.Rollback() will handle reverting changes
		log.Fatalf("Failed to insert Dave: %v", err)
	}

	_, err = tx.Exec(insertStmt, "Eve", 22)
	if err != nil {
		log.Fatalf("Failed to insert Eve: %v", err)
	}

	// 3. Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	fmt.Println("Transaction committed successfully. Records inserted.")
}
