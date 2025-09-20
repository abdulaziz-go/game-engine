package server

import (
	"log"
	"net"
)

func StartServer() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
	defer listener.Close()

	log.Println("Tic-Tac-Toe server listening on :8080")

	gameServer := NewTicTacToeServer()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go gameServer.handleClient(conn)
	}
}
