package panicrecover

import "fmt"

func SafeDivision(a, b int) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Recovered from panic:", err)
		}
	}()

	result := a / b
	fmt.Printf("%d / %d = %d\n", a, b, result)
}

func DangerousOperation() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Caught panic: %v\n", err)
		}
	}()

	s := []int{1, 2, 3}
	fmt.Println(s[10]) // panic: index out of range
}
