package main

import (
	"fmt"
	"time"

	"github.com/paulbellamy/ratecounter"
)

func main() {
	errorRate := time.Tick(1 * time.Second)
	checkErrorRate := time.Tick(10 * time.Second)
	// stop := time.After(15 * time.Second)

	counter := ratecounter.NewRateCounter(60 * time.Second)

	q := make(chan bool)
	go foo(q)

	for {
		select {
		case <-errorRate:
			counter.Incr(1)
		case <-checkErrorRate:
			fmt.Printf("errors per min: %v\n", counter.Rate())
			if counter.Rate() > 1 {
				q <- true
			}
		}
	}

}

func foo(q chan bool) {
	for {
		select {
		case <-q:
			fmt.Println("END")
			return
		default:
			fmt.Printf("work ....\n")
			time.Sleep(1 * time.Second)
		}
	}
}
