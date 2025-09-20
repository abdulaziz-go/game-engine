package main

import (
	"fmt"
	"game/client"
	"game/server"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go [server|client]")
		return
	}

	switch os.Args[1] {
	case "server":
		log.Println("Starting Tic-Tac-Toe server...")
		server.StartServer()
	case "client":
		log.Println("Starting Tic-Tac-Toe client...")
		client.StartClient()
	default:
		fmt.Println("Usage: go run main.go [server|client]")
	}
}
