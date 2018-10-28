package main

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
)

func main() {

	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	host, _ := os.Hostname()
	os := runtime.GOOS

	// Current User
	fmt.Println("Hi " + user.Name + " (id: " + user.Uid + ")")
	fmt.Println("Username: " + user.Username)
	fmt.Println("Home Dir: " + user.HomeDir)
	fmt.Println("Host: " + host)
	fmt.Println("OS: " + os)
}
