package stl

import (
	"fmt"
	"slices"
)

func SliceExamples() {
	// Example 1: Contains - check if element exists
	nums := []int{1, 2, 3, 4, 5}
	fmt.Println("Example 1: ", slices.Contains(nums, 3))
	fmt.Println("Example 1: ", slices.Contains(nums, 10))

	// Example 2: Index - find element's position
	idx := slices.Index(nums, 4)
	fmt.Println("Example 2: ", idx)

	// Example 3: Sort - sort slice
	unsorted := []int{5, 2, 8, 1, 9}
	slices.Sort(unsorted)
	fmt.Println("Example 3: ", unsorted)

	// Example 4: Reverse - reverse slice
	items := []string{"a", "b", "c", "d"}
	slices.Reverse(items)
	fmt.Println("Example 4: ", items)

	// Example 5: Clone - create copy of slice
	original := []int{1, 2, 3}
	copied := slices.Clone(original)
	fmt.Println("Example 5: ", copied)

	// Example 6: Delete - remove elements
	slice := []int{10, 20, 30, 40, 50}
	slice = slices.Delete(slice, 1, 3)
	fmt.Println("Example 6: ", slice)

	// Example 7: Compact - remove consecutive duplicates
	dups := []int{1, 1, 2, 2, 2, 3, 3, 4}
	dups = slices.Compact(dups)
	fmt.Println("Example 7: ", dups)

	// Example 8: Equal - compare slices
	a := []int{1, 2, 3}
	b := []int{1, 2, 3}
	fmt.Println("Example 8: ", slices.Equal(a, b))
}
