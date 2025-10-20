package stl

import (
	"fmt"
	"sort"
)

func SortExamples() {
	// Example 1: Sort integers
	numbers := []int{5, 2, 8, 1, 9}
	sort.Ints(numbers)
	fmt.Println("Sorted integers:", numbers)

	// Example 2: Sort strings
	words := []string{"banana", "apple", "cherry", "date"}
	sort.Strings(words)
	fmt.Println("Sorted strings:", words)

	// Example 3: Sort floats
	floats := []float64{3.14, 1.41, 2.71, 1.73}
	sort.Float64s(floats)
	fmt.Println("Sorted floats:", floats)

	// Example 4: Check if already sorted
	fmt.Println("Is [1,2,3] sorted?", sort.IntsAreSorted([]int{1, 2, 3}))
	fmt.Println("Is [3,1,2] sorted?", sort.IntsAreSorted([]int{3, 1, 2}))

	// Example 5: Custom sort with interface. Using the same Person struct defined before in the CMP examples
	people := []Person{
		{"Alice", 30},
		{"Bob", 25},
		{"Charlie", 35},
	}

	// Sort by age
	sort.Slice(people, func(i, j int) bool {
		return people[i].Age < people[j].Age
	})
	fmt.Println("Sorted by age:", people)

	// Sort by name (reverse)
	sort.Slice(people, func(i, j int) bool {
		return people[i].Name > people[j].Name
	})
	fmt.Println("Sorted by name (reverse):", people)
}
