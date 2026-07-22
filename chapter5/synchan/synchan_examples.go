package synchan

import (
	"fmt"
	"time"
)

func sender(ch chan string) {
	fmt.Println("Sender: Preparing to send message...")
	time.Sleep(2 * time.Second)

	fmt.Println("Sender: Sending message (will block until receiver is ready)...")
	ch <- "Hello from sender!"

	fmt.Println("Sender: Message sent successfully!")
}

func receiver(ch chan string) {
	fmt.Println("Receiver: Doing some work before receiving...")
	time.Sleep(3 * time.Second)

	fmt.Println("Receiver: Ready to receive message...")
	msg := <-ch

	fmt.Println("Receiver: Received message:", msg)
}

func SyncChanExample() {
	// Create an unbuffered channel for syncronization
	syncChannel := make(chan string)

	fmt.Println("Main goroutine: Starting goroutines with synchronous channel...")
	fmt.Println("---")

	// Launch sender and receiver goroutines
	go sender(syncChannel)
	go receiver(syncChannel)

	// Wait for goroutines to complete
	time.Sleep(5 * time.Second)

	fmt.Println("---")
	fmt.Println("Main goroutine: Program completed!")
}
