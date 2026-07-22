package fsystem

import (
	"bufio"
	"fmt"
	"os"
)

func FileSystemExamples() {
	filename := "example.txt"

	// Example 1: Create and write to a file
	fmt.Println("Example 1: Creating and writing to file...")
	err := os.WriteFile(filename, []byte("Hello, World!\nThis is line 2.\n"), 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}
	fmt.Println("✓ File created and written")

	// Example 2: Read the entire file
	fmt.Println("\nExample 2: Reading entire file...")
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	fmt.Printf("Content:\n%s", content)

	// Example 3: Append to file
	fmt.Println("Example 3: Appending to file...")
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString("This line was appended.\n")
	if err != nil {
		fmt.Println("Error appending:", err)
		return
	}
	fmt.Println("✓ Text appended")

	// Example 4: Read file line by line
	fmt.Println("\nExample 4: Reading file line by line...")
	file, err = os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 1
	for scanner.Scan() {
		fmt.Printf("Line %d: %s\n", lineNum, scanner.Text())
		lineNum++
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error scanning file:", err)
	}

	// Example 5: Check if file exists
	fmt.Println("\nExample 5: Checking file existence...")
	if _, err := os.Stat(filename); err == nil {
		fmt.Println("✓ File exists")
	} else if os.IsNotExist(err) {
		fmt.Println("✗ File does not exist")
	}

	// Example 6: Get file info
	fmt.Println("\nExample 6: Getting file info...")
	fileInfo, err := os.Stat(filename)
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return
	}
	fmt.Printf("Name: %s\n", fileInfo.Name())
	fmt.Printf("Size: %d bytes\n", fileInfo.Size())
	fmt.Printf("Mode: %s\n", fileInfo.Mode())
	fmt.Printf("Modified: %s\n", fileInfo.ModTime())

	// Example 7: Delete file
	fmt.Println("\nExample 7: Deleting file...")
	err = os.Remove(filename)
	if err != nil {
		fmt.Println("Error deleting file:", err)
		return
	}
	fmt.Println("✓ File deleted")
}
