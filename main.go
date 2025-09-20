package main

import (
	"awesomeProject1/client"
	"awesomeProject1/server"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  go run main.go server")
		fmt.Println("  go run main.go client")
		return
	}

	switch os.Args[1] {
	case "server":
		server.Start()
	case "client":
		client.Start()
	default:
		fmt.Println("Use 'server' or 'client'")
	}
}
