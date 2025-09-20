package client

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"game/protocol"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

type GameClient struct {
	conn       net.Conn
	playerID   uint32
	playerName string
	symbol     uint8
}

func StartClient() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal("Error connecting to server:", err)
	}
	defer conn.Close()

	client := &GameClient{conn: conn}

	// Get player name
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	client.playerName = name

	// Join game
	player := protocol.Player{Name: name}
	playerData := protocol.EncodePlayer(player)
	joinMsg := protocol.EncodeTLV(protocol.MSG_PLAYER_JOIN, playerData)
	client.conn.Write(joinMsg)

	// Start listening for server messages
	go client.listenForMessages()

	// Handle user input
	client.handleUserInput(reader)
}

func (gc *GameClient) listenForMessages() {
	for {
		header := make([]byte, 3)
		_, err := gc.conn.Read(header)
		if err != nil {
			break
		}

		length := binary.BigEndian.Uint16(header[1:3])
		data := make([]byte, length)

		if length > 0 {
			_, err = gc.conn.Read(data)
			if err != nil {
				break
			}
		}

		fullMsg := append(header, data...)
		msg, err := protocol.DecodeTLV(fullMsg)
		if err != nil {
			continue
		}

		switch msg.Type {
		case protocol.MSG_WAITING:
			fmt.Println(string(msg.Value))

		case protocol.MSG_GAME_START:
			fmt.Println("\n" + string(msg.Value))
			fmt.Println("You are playing as:", map[uint8]string{protocol.X: "X", protocol.O: "O"}[gc.symbol])

		case protocol.MSG_GAME_STATE:
			board, err := protocol.DecodeGameBoard(msg.Value)
			if err != nil {
				continue
			}
			gc.displayBoard(board)

		case protocol.MSG_ERROR:
			fmt.Println("Error:", string(msg.Value))
			return
		}
	}
}

func (gc *GameClient) displayBoard(board protocol.GameBoard) {
	fmt.Print("\033[H\033[2J") // Clear screen
	fmt.Println("=== TIC-TAC-TOE ===")
	fmt.Printf("Player: %s\n", gc.playerName)

	symbols := map[uint8]string{protocol.EMPTY: " ", protocol.X: "X", protocol.O: "O"}

	fmt.Println("   1   2   3")
	for i := 0; i < 3; i++ {
		fmt.Printf("%d  %s | %s | %s \n", i+1,
			symbols[board.Board[i*3]],
			symbols[board.Board[i*3+1]],
			symbols[board.Board[i*3+2]])
		if i < 2 {
			fmt.Println("  -----------")
		}
	}

	if board.GameStatus == 1 {
		if board.Winner == 3 {
			fmt.Println("\n🤝 It's a draw!")
		} else {
			winner := symbols[board.Winner]
			fmt.Printf("\n🎉 Player %s wins!\n", winner)
		}
		fmt.Println("Game over! Type 'quit' to exit.")
	} else {
		currentPlayer := symbols[board.CurrentTurn]
		fmt.Printf("\nCurrent turn: %s\n", currentPlayer)
		if board.CurrentTurn == gc.symbol {
			fmt.Println("It's YOUR turn! Enter position (row col): ")
		} else {
			fmt.Println("Waiting for opponent's move...")
		}
	}

	fmt.Print("> ")
}

func (gc *GameClient) handleUserInput(reader *bufio.Reader) {
	fmt.Println("Connected! Wait for the game to start...")

	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "quit" {
			break
		}

		parts := strings.Split(input, " ")
		if len(parts) == 2 {
			row, err1 := strconv.Atoi(parts[0])
			col, err2 := strconv.Atoi(parts[1])

			if err1 == nil && err2 == nil && row >= 1 && row <= 3 && col >= 1 && col <= 3 {
				position := uint8((row-1)*3 + (col - 1))
				move := protocol.Move{PlayerID: gc.playerID, Position: position}
				moveData := protocol.EncodeMove(move)
				moveMsg := protocol.EncodeTLV(protocol.MSG_MAKE_MOVE, moveData)
				gc.conn.Write(moveMsg)
			} else {
				fmt.Println("Invalid input! Use: row col (e.g., '2 1' for row 2, column 1)")
				fmt.Print("> ")
			}
		} else if input != "" {
			fmt.Println("Enter position as: row col (e.g., '1 1' for top-left)")
			fmt.Print("> ")
		}
	}
}
