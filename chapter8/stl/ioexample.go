package stl

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func IOExamples() {
	// Example 1: Copy from reader to writer
	fmt.Println("=== Example 1: io.Copy ===")
	src := strings.NewReader("Hello, World!")
	io.Copy(os.Stdout, src)
	fmt.Println()

	// Example 2: LimitReader - limit the number of bytes read
	fmt.Println("\n=== Example 2: io.LimitReader ===")
	reader := strings.NewReader("This is a longer text")
	limitedReader := io.LimitReader(reader, 7)
	io.Copy(os.Stdout, limitedReader)
	fmt.Println()

	// Example 3: MultiReader - combine multiple readers
	fmt.Println("\n=== Example 3: io.MultiReader ===")
	r1 := strings.NewReader("Hello, ")
	r2 := strings.NewReader("World!")
	multiReader := io.MultiReader(r1, r2)
	io.Copy(os.Stdout, multiReader)
	fmt.Println()

	// Example 4: TeeReader - read and write simultaneously
	fmt.Println("\n=== Example 4: io.TeeReader ===")
	src = strings.NewReader("Tee example")
	var buf strings.Builder
	teeReader := io.TeeReader(src, &buf)
	io.Copy(os.Stdout, teeReader)
	fmt.Printf("\nBuffer contains: %s\n", buf.String())

	// Example 5: ReadAll - read all data from reader
	fmt.Println("\n=== Example 5: io.ReadAll ===")
	reader = strings.NewReader("Read everything")
	data, _ := io.ReadAll(reader)
	fmt.Printf("Read %d bytes: %s\n", len(data), string(data))
}
