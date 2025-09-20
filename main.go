package main

import (
	"game/client"
	"game/server"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		panic("example running: go run main.go server|client")
	}

	switch os.Args[1] {
	case "server":
		server.Start()
	case "client":
		client.Start()
	}
}
