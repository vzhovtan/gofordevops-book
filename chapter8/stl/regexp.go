package stl

import (
	"fmt"
	"regexp"
)

func RegexExamples() {
	// Example 1. Basic pattern matching
	re := regexp.MustCompile(`hello`)
	fmt.Println("Example 1")
	fmt.Println(re.MatchString("hello world"))   // true
	fmt.Println(re.MatchString("goodbye world")) // false

	// Example 2. Find first match
	fmt.Println("Example 2")
	re = regexp.MustCompile(`\d+`)
	match := re.FindString("I have 42 apples and 10 oranges")
	fmt.Println("First number:", match) // 42

	// Example 3. Find all matches
	fmt.Println("Example 3")
	matches := re.FindAllString("I have 42 apples and 10 oranges", -1)
	fmt.Println("All numbers:", matches) // [42 10]

	// Example 4. Replace with pattern
	fmt.Println("Example 4")
	re = regexp.MustCompile(`\s+`)
	result := re.ReplaceAllString("hello   world  test", " ")
	fmt.Println("Normalized:", result) // hello world test

	// Example 5. Extract groups (submatches)
	fmt.Println("Example 5")
	re = regexp.MustCompile(`(\w+)@(\w+\.\w+)`)
	email := "john@example.com"
	groups := re.FindStringSubmatch(email)
	if len(groups) > 0 {
		fmt.Println("User:", groups[1])   // john
		fmt.Println("Domain:", groups[2]) // example.com
	}

	// Example 6. Case-insensitive matching
	fmt.Println("Example 6")
	re = regexp.MustCompile(`(?i)hello`)
	fmt.Println(re.MatchString("HELLO")) // true
	fmt.Println(re.MatchString("Hello")) // true
}
