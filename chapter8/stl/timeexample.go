package stl

import (
	"fmt"
	"time"
)

func TimeExamples() {
	// Example1: Get current time
	now := time.Now()
	fmt.Println("Example1: Current time:", now)

	// Example2: Format time
	formatted := now.Format("2006-01-02 15:04:05")
	fmt.Println("Example2: Formatted:", formatted)

	// Example3: Parse time from string
	parsed, _ := time.Parse("2006-01-02", "2025-10-19")
	fmt.Println("Example13: Parsed:", parsed)

	// Example4: Add/subtract duration
	tomorrow := now.Add(24 * time.Hour)
	yesterday := now.Add(-24 * time.Hour)
	fmt.Println("Example4: Tomorrow:", tomorrow.Format("2006-01-02"))
	fmt.Println("Example4: Yesterday:", yesterday.Format("2006-01-02"))

	// Example5: Calculate difference
	diff := tomorrow.Sub(now)
	fmt.Println("Example5: Difference:", diff)

	// Example6: Sleep
	fmt.Println("Example6: Sleeping for 1 second...")
	time.Sleep(1 * time.Second)
	fmt.Println("Done!")

	// Example7: Create a timer
	timer := time.NewTimer(2 * time.Second)
	<-timer.C
	fmt.Println("Example7: Timer finished!")

	// Example8: Ticker example
	ticker := time.NewTicker(500 * time.Millisecond)
	go func() {
		for i := 0; i < 3; i++ {
			<-ticker.C
			fmt.Println("Example8: Tick")
		}
		ticker.Stop()
	}()
	time.Sleep(2 * time.Second)
}
