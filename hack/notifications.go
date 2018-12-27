package main

import (
	"log"
	"math/rand"
	"time"
)

var doneChan = make(chan bool)
var keepAliveChan = make(chan bool)

func main() {
	r := rand.Intn(8) + 1
	ticker := time.NewTicker(time.Duration(r) * time.Second)
	quit := make(chan bool)

	go notificationManager()
	go func() {
		for {
			select {
			case <-doneChan:
				ticker.Stop()
				log.Println("nothing new from ticker ...")
				quit <- true
				return
			}
		}
	}()

	for {
		select {
		case <-ticker.C:
			ticker.Stop()
			r := rand.Intn(8) + 1
			ticker = time.NewTicker(time.Duration(time.Duration(r) * time.Second))
			log.Printf("tick in %d\n", r)
			callback()
		case <-quit:
			ticker.Stop()
			log.Println("..ticker stopped!")
			return
		}
	}
}

func callback() {
	log.Println("callback(): work in progress ...")
	keepAliveChan <- true
}

func notificationManager() {
	for {
		select {
		case <-keepAliveChan:
			log.Println("working ...")
		case <-time.After(time.Duration(10 * time.Second)):
			log.Println("done!")
			doneChan <- true
		}
	}
}
