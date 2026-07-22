package stl

import (
	"fmt"
	"strings"
)

func StringsExamples() {
	// Example 1: Contains - check if substring exists
	text := "Hello, World!"
	fmt.Println(strings.Contains(text, "World"))
	fmt.Println(strings.Contains(text, "xyz"))

	// Example 2: Index - find position of substring
	fmt.Println(strings.Index(text, "World"))
	fmt.Println(strings.Index(text, "xyz"))

	// Example 3: Split - break string into slice
	csv := "apple,banana,orange"
	fruits := strings.Split(csv, ",")
	fmt.Println(fruits)

	// Example 4: Join - combine slice into string
	joined := strings.Join(fruits, "; ")
	fmt.Println(joined)

	// Example 5: ToUpper/ToLower - case conversion
	fmt.Println(strings.ToUpper("hello"))
	fmt.Println(strings.ToLower("WORLD"))

	// Example 6: Replace - substitute substrings
	old := "cat and cat"
	replaced := strings.Replace(old, "cat", "dog", 1)
	fmt.Println(replaced)

	replaceAll := strings.ReplaceAll(old, "cat", "dog")
	fmt.Println(replaceAll)

	// Example 7: TrimSpace - remove leading/trailing whitespace
	padded := "  hello world  "
	fmt.Println(strings.TrimSpace(padded))

	// Example 8: Repeat - duplicate string
	fmt.Println(strings.Repeat("ab", 3))

	// Example 9: Count - count occurrences
	fmt.Println(strings.Count("banana", "a"))

	// Example 10: HasPrefix/HasSuffix - check start/end
	fmt.Println(strings.HasPrefix("hello.go", ".go"))
	fmt.Println(strings.HasSuffix("hello.go", ".go"))
}
