package mtx

import (
	"fmt"
	"sync"
)

type Counter struct {
	mu    sync.Mutex
	value int
}

func (c *Counter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}

func (c *Counter) GetValue() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

func MutexPlay() {
	counter := &Counter{}
	var wg sync.WaitGroup

	// Launch 1000 goroutines that each increment the counter
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Increment()
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Print final value
	fmt.Printf("Final counter value: %d\n", counter.GetValue())
	fmt.Println("Expected: 1000")
}
