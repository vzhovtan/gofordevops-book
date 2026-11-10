package send_receive

import (
	"fmt"
	"time"
)

// dataProducer takes a send-only channel.
// It can only send messages to the 'messages' channel.
func dataProducer(messages chan<- string) {
	defer close(messages) // Close the channel when done sending
	for i := 0; i < 5; i++ {
		msg := fmt.Sprintf("Message %d", i)
		fmt.Println("Producer: Sending", msg)
		messages <- msg
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Println("Producer: Finished sending")
}

// dataConsumer takes a receive-only channel.
// It can only receive messages from the 'messages' channel.
func dataConsumer(messages <-chan string, done chan<- bool) {
	fmt.Println("Consumer: Waiting for messages...")
	for msg := range messages {
		fmt.Println("Consumer: Received", msg)
	}
	fmt.Println("Consumer: Channel closed, exiting")
	done <- true // Signal completion
}

// dataRelay takes a receive-only channel (input) and a send-only channel (output).
// It reads from 'input' and sends to 'output'.
func dataRelay(input <-chan string, output chan<- string) {
	defer close(output)
	for msg := range input {
		relayedMsg := fmt.Sprintf("Relayed - %s", msg)
		fmt.Printf("Relay: Relaying '%s'\n", msg)
		output <- relayedMsg
	}
	fmt.Println("Relay: Input channel closed")
}

func SendReceive() {
	// Example 1: Basic Producer-Consumer
	fmt.Println("--- Example 1: Producer-Consumer ---")
	messages1 := make(chan string)
	done := make(chan bool)

	go dataProducer(messages1)       // Pass the channel, will be treated as send-only
	go dataConsumer(messages1, done) // Pass the channel, will be treated as receive-only

	<-done // Wait for the consumer to finish
	fmt.Println("--- Example 1 Finished ---")

	// Example 2: Producer -> Relay -> Consumer
	fmt.Println("\n--- Example 2: Producer -> Relay -> Consumer ---")
	messages2 := make(chan string)
	relayedMessages := make(chan string)
	done2 := make(chan bool)

	go dataProducer(messages2)               // Producer sends to messages2
	go dataRelay(messages2, relayedMessages) // Relay reads from messages2, sends to relayedMessages
	go dataConsumer(relayedMessages, done2)  // Consumer reads from relayedMessages

	<-done2 // Wait for the final consumer to finish
	fmt.Println("--- Example 2 Finished ---")
}
