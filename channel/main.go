package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan int, 10)
	go receiveCh(ch)

	for i := 0; i < 10; i++ {
		ch <- i
		time.Sleep(1 * time.Second)
	}

	close(ch)
	time.Sleep(1 * time.Second)
}

func receiveCh(c chan int) {
	for v := range c {
		fmt.Println("v: ", v)
	}
	fmt.Printf("receive ch quit")
}
