package signals

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func RunAndWait() {
	// Create a channel to receive OS signals
	sigChan := make(chan os.Signal, 1)

	// Register the channel to receive specific signals
	// SIGINT is triggered by Ctrl+C
	signal.Notify(sigChan, syscall.SIGINT)

	// Create a channel to signal when cleanup is done
	done := make(chan bool, 1)

	// Start a goroutine to simulate some work
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				fmt.Println("Processing some things...")
				time.Sleep(2 * time.Second)
			}
		}
	}()

	fmt.Println("Program started. Press Ctrl+C to stop.")

	// Block until a signal is received
	sig := <-sigChan
	fmt.Printf("\nReceived signal: %v\n", sig)

	// Perform cleanup operations
	fmt.Println("Performing cleanup...")
	done <- true
	time.Sleep(1 * time.Second)

	fmt.Println("Cleanup complete. Exiting.")
}
