package wg

import (
	"fmt"
	"sync"
	"time"
)

func worker(id int, wg *sync.WaitGroup) {
	// Notify WaitGroup when this function completes
	defer wg.Done()

	fmt.Printf("Worker %d: Starting task\n", id)

	// Simulate some work with sleep
	time.Sleep(time.Duration(id) * time.Second)

	fmt.Printf("Worker %d: Task completed\n", id)
}

func WgExample(numWorkers int) {
	// Create a WaitGroup
	var wg sync.WaitGroup

	fmt.Println("Starting workers...")

	// Launch multiple concurrent goroutines
	for i := 1; i <= numWorkers; i++ {
		// Increment the WaitGroup counter
		wg.Add(1)

		// Launch goroutine
		go worker(i, &wg)
	}

	// Wait for all goroutines to complete
	fmt.Println("Waiting for all workers to finish...")
	wg.Wait()

	fmt.Println("All workers completed! Program finished.")
}
