package deadlock

import "fmt"

func Deadlock() {
	ch := make(chan int)

	// Trying to send on an unbuffered channel without a receiver
	fmt.Println("Attempting to send on channel...")
	ch <- 42 // This will block forever - DEADLOCK!

	fmt.Println("This line will never execute")
}
