package buffered

import (
	"fmt"
	"time"
)

func BufferedComms(capacity int) {
	// Create a buffered channel with capacity of 3 and element type of string
	ch := make(chan string, capacity)

	// Sending goroutine - sends values to the channel
	go func() {
		messages := []string{"Hello", "World", "From", "Buffered", "Channel"}

		for i, msg := range messages {
			fmt.Printf("Producer: Sending message %d: %s\n", i+1, msg)
			ch <- msg // Will block only when buffer is full
			fmt.Printf("Producer: Message %d sent (buffer has space)\n", i+1)
			time.Sleep(500 * time.Millisecond)
		}
		close(ch) // Close channel when done sending
		fmt.Println("Producer: Channel closed")
	}()

	// Receiver goroutine - receives values from the channel
	go func() {
		time.Sleep(2 * time.Second) // Delay to show buffering effect
		fmt.Println("Consumer: Starting to receive messages...")

		for msg := range ch {
			fmt.Printf("Consumer: Received message: %s\n", msg)
			time.Sleep(1 * time.Second) // Slow consumer
		}
		fmt.Println("Consumer: All messages received")
	}()

	// Keep main alive
	time.Sleep(10 * time.Second)
	fmt.Println("Program finished")
}
