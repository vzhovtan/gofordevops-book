package stl

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func BufioExamples() {
	// Example 1: Reading from standard input
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter your name: ")
	name, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	name = strings.TrimSpace(name)
	fmt.Printf("Hello, %s!\n\n", name)

	// Example 2: Reading a file line by line
	file, err := os.Open("test.txt")
	if err != nil {
		fmt.Println("Creating test.txt...")
		file, err = os.Create("test.txt")
		if err != nil {
			log.Fatal(err)
		}
		file.WriteString("Test1\nTest2\nTest3\n")
		file.Close()
		file, err = os.Open("test.txt")
		if err != nil {
			log.Fatal(err)
		}
	}
	defer file.Close()

	fmt.Println("File contents:")
	scanner := bufio.NewScanner(file)
	lineNum := 1
	for scanner.Scan() {
		fmt.Printf("%d: %s\n", lineNum, scanner.Text())
		lineNum++
	}

	// Example 3: Writing with bufio
	writer := bufio.NewWriter(os.Stdout)
	fmt.Fprintf(writer, "\nBuffered output: This is line 1\n")
	fmt.Fprintf(writer, "Buffered output: This is line 2\n")
	writer.Flush() // It has to be flushed
}
