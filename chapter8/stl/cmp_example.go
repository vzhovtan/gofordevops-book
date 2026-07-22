package stl

import (
	"cmp"
	"fmt"
)

type Person struct {
	Name string
	Age  int
}

func ComparePersons(a, b Person) int {
	if result := cmp.Compare(a.Name, b.Name); result != 0 {
		return result
	}
	return cmp.Compare(a.Age, b.Age)
}

func CmpExamples() {
	// Example 1: Compare integers
	fmt.Println("Integer comparisons:")
	fmt.Println(cmp.Compare(5, 10))  // Output: -1 (5 < 10)
	fmt.Println(cmp.Compare(10, 10)) // Output: 0 (equal)
	fmt.Println(cmp.Compare(15, 10)) // Output: 1 (15 > 10)

	// Example 2: Compare strings
	fmt.Println("\nString comparisons:")
	fmt.Println(cmp.Compare("apple", "banana")) // Output: -1
	fmt.Println(cmp.Compare("hello", "hello"))  // Output: 0
	fmt.Println(cmp.Compare("zebra", "apple"))  // Output: 1

	// Example 3:Using Or for multiple comparisons
	fmt.Println("\nUsing Or function:")
	a, b := 5, 5
	c, d := 10, 3
	result := cmp.Or(
		cmp.Compare(a, b),
		cmp.Compare(c, d),
	)
	fmt.Printf("Or(Compare(%d,%d), Compare(%d,%d)) = %d\n", a, b, c, d, result)
	// The first comparison is 0, so moves to second: Compare(10, 3) = 1

	// Example 4: Compare structs
	p1 := Person{Name: "Alice", Age: 30}
	p2 := Person{Name: "Bob", Age: 25}
	fmt.Printf("ComparePersons(p1, p2) = %d (p1 > p2)\n", ComparePersons(p1, p2))

}
