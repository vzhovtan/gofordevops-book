package unbuffered

import "fmt"

func UnbufferedChannel() {
	// Create an unbuffered channel
	ch := make(chan int)

	// Launch a goroutine to send data
	go func() {
		fmt.Println("Sender: Preparing to send value 42")
		ch <- 42 // This will block until receiver is ready
		fmt.Println("Sender: Value sent successfully")
	}()

	// Launch a goroutine to receive data
	go func() {
		fmt.Println("Receiver: Waiting to receive value")
		value := <-ch // This will block until sender sends
		fmt.Println("Receiver: Received value:", value)
	}()

	// Avoid completion of main goroutine to let other goroutines complete
	var input string
	fmt.Scanln(&input)
}
