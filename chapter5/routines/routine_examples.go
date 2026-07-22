package routines

import (
	"fmt"
	"math/rand"
	"time"
)

func f(n int) {
	for i := 0; i < 10; i++ {
		fmt.Println(n, ":", i)
		amt := time.Duration(rand.Intn(250))
		time.Sleep(time.Millisecond * amt)
	}
}

func Routine() {
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			go f(i)
			go f(j)
		}
	}
	var input string
	fmt.Scanln(&input)
}
