package main

import (
	"fmt"
	"time"
)

func main() {
	serverChan := make(chan chan string)
	go pub(serverChan)

	client1Chan := make(chan string)
	client2Chan := make(chan string)
	serverChan <- client1Chan
	serverChan <- client2Chan

	go sub1(client1Chan)
	go sub2(client2Chan)

	time.Sleep(10 * time.Second)
}

func pub(serverChan chan chan string) {
	var clients []chan string
	for {
		select {
		case client, _ := <-serverChan:
			clients = append(clients, client)
		case <-time.After(time.Second):
			// Broadcast the number of clients to all clients:
			for _, c := range clients {
				c <- "Ping"
			}
		}
	}
}

func sub1(clientChan chan string) {
	for {
		text, _ := <-clientChan
		fmt.Printf(">>> 1: %s\n", text)
	}
}

func sub2(clientChan chan string) {
	for {
		text, _ := <-clientChan
		fmt.Printf(">>> 2: %s\n", text)
	}
}
