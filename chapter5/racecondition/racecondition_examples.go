package racecondition

import (
	"fmt"
	"sync"
)

func RaceCondition() {
	balance := 0
	var wg sync.WaitGroup

	// Simulate 1000 deposits of $1 each
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// RACE CONDITION: Multiple goroutines reading and writing balance
			currentBalance := balance
			currentBalance += 1
			balance = currentBalance
		}()
	}

	wg.Wait()
	fmt.Println("Final balance:", balance)
	fmt.Println("Expected balance: 1000")
	fmt.Println("\nRun with: go run -race main.go to detect the race condition")
}
