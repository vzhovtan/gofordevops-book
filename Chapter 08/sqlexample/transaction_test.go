package sqlexample

import (
	"database/sql"
	"os"
	"testing"

	_ "modernc.org/sqlite"
)

func TestInsertPeopleTransaction(t *testing.T) {
	// 1. Cleanup the database file after test finishes
	defer os.Remove("test.db")

	// 2. Pre-setup: Create the table since InsertPeopleTransaction assumes it exists
	db, err := sql.Open("sqlite", "test.db")
	if err != nil {
		t.Fatalf("sql.Open() failed: %v", err)
	}
	defer db.Close()

	createTable := `
CREATE TABLE IF NOT EXISTS people (
id INTEGER PRIMARY KEY AUTOINCREMENT,
name TEXT NOT NULL,
age INTEGER NOT NULL
);`
	if _, err := db.Exec(createTable); err != nil {
		t.Fatalf("db.Exec() failed to create table: %v", err)
	}

	// 3. Action: Run the transaction function
	InsertPeopleTransaction()

	// 4. Assertion: Verify both users were inserted
	rows, err := db.Query("SELECT name, age FROM people ORDER BY name ASC")
	if err != nil {
		t.Fatalf("db.Query() failed: %v", err)
	}
	defer rows.Close()

	type personRecord struct {
		Name string
		Age  int
	}
	var got []personRecord

	for rows.Next() {
		var p personRecord
		if err := rows.Scan(&p.Name, &p.Age); err != nil {
			t.Fatalf("rows.Scan() failed: %v", err)
		}
		got = append(got, p)
	}

	want := []personRecord{
		{Name: "Dave", Age: 40},
		{Name: "Eve", Age: 22},
	}

	if len(got) != len(want) {
		t.Fatalf("Expected %d records, got %d", len(want), len(got))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("At index %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}
