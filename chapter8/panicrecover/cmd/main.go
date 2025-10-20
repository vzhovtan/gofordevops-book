package main

import (
	"fmt"
	"panicrecover"
)

func main() {
	panicrecover.FileOpenRead("nonexistingfile.txt")

	fmt.Println("Example 1: Safe division")
	panicrecover.SafeDivision(10, 2)

	fmt.Println("\nExample 2: Division by zero")
	// Will panic and recover
	panicrecover.SafeDivision(10, 0)

	fmt.Println("\nExample 3: Index out of range")
	// Will panic and recover
	panicrecover.DangerousOperation()

	fmt.Println("\nProgram continues normally")
}
