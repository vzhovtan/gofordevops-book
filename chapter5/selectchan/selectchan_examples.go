package selectchan

import (
	"fmt"
	"time"
)

func SelectExample() {
	// Create two channels
	channel1 := make(chan string)
	channel2 := make(chan string)

	// Goroutine 1: sends to channel1 after 1 second
	go func() {
		time.Sleep(1 * time.Second)
		channel1 <- "Message from Channel 1"
	}()

	// Goroutine 2: sends to channel2 after 2 seconds
	go func() {
		time.Sleep(2 * time.Second)
		channel2 <- "Message from Channel 2"
	}()

	// Using select to receive from multiple channels
	fmt.Println("Waiting for messages...")

	select {
	case msg1 := <-channel1:
		fmt.Println("Received:", msg1)
	case msg2 := <-channel2:
		fmt.Println("Received:", msg2)
	case <-time.After(5 * time.Second):
		fmt.Println("Timeout: No message received")
	}

	fmt.Println("Program completed!")
}
