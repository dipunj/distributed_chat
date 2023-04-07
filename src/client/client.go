package main

import (
	"chat/client/state"
	"chat/client/ui"
	"os"
)

func main() {

	// todo
	// 1. wait for the connection to be established before starting the user input loop
	// 2. add a mechanism to know if the connection is broken
	// 3. configure timeout

	os.Stdout.WriteString("Welcome...enter h for help")

	state.Wait.Add(1)
	defer state.Wait.Done()
	go ui.UserViewLoop()

	// on the main thread
	ui.UserInputLoop()
}
