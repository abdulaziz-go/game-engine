package server

import (
	"encoding/binary"
	"game/protocol"
	"log"
	"net"
	"sync"
)

type TicTacToeServer struct {
	players    map[uint32]protocol.Player
	clients    map[uint32]net.Conn
	gameBoard  protocol.GameBoard
	playerID   uint32
	gameActive bool
	mutex      sync.RWMutex
}

func NewTicTacToeServer() *TicTacToeServer {
	return &TicTacToeServer{
		players:   make(map[uint32]protocol.Player),
		clients:   make(map[uint32]net.Conn),
		playerID:  1,
		gameBoard: protocol.GameBoard{CurrentTurn: protocol.X},
	}
}

func (tts *TicTacToeServer) handleClient(conn net.Conn) {
	defer conn.Close()

	var currentPlayerID uint32

	for {
		header := make([]byte, 3)
		_, err := conn.Read(header)
		if err != nil {
			break
		}

		length := binary.BigEndian.Uint16(header[1:3])
		data := make([]byte, length)

		if length > 0 {
			_, err = conn.Read(data)
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
		case protocol.MSG_PLAYER_JOIN:
			player, err := protocol.DecodePlayer(msg.Value)
			if err != nil {
				continue
			}

			tts.mutex.Lock()
			if len(tts.players) >= 2 {
				// Game is full
				errorMsg := protocol.EncodeTLV(protocol.MSG_ERROR, []byte("Game is full"))
				conn.Write(errorMsg)
				tts.mutex.Unlock()
				return
			}

			player.ID = tts.playerID
			currentPlayerID = tts.playerID

			// Assign X or O
			if len(tts.players) == 0 {
				player.Symbol = protocol.X
			} else {
				player.Symbol = protocol.O
			}

			tts.players[player.ID] = player
			tts.clients[player.ID] = conn
			tts.playerID++

			log.Printf("Player %s joined as %s", player.Name,
				map[uint8]string{protocol.X: "X", protocol.O: "O"}[player.Symbol])

			if len(tts.players) == 2 {
				tts.gameActive = true
				tts.broadcastGameStart()
			} else {
				// Send waiting message
				waitMsg := protocol.EncodeTLV(protocol.MSG_WAITING, []byte("Waiting for another player..."))
				conn.Write(waitMsg)
			}
			tts.mutex.Unlock()

		case protocol.MSG_MAKE_MOVE:
			if !tts.gameActive {
				continue
			}

			move, err := protocol.DecodeMove(msg.Value)
			if err != nil {
				continue
			}

			tts.mutex.Lock()
			player, exists := tts.players[currentPlayerID]
			if !exists || player.Symbol != tts.gameBoard.CurrentTurn {
				tts.mutex.Unlock()
				continue
			}

			// Validate move
			if move.Position > 8 || tts.gameBoard.Board[move.Position] != protocol.EMPTY {
				tts.mutex.Unlock()
				continue
			}

			// Make the move
			tts.gameBoard.Board[move.Position] = player.Symbol

			// Check for win or draw
			winner := tts.checkWinner()
			if winner != protocol.EMPTY {
				tts.gameBoard.GameStatus = 1
				tts.gameBoard.Winner = winner
				tts.gameActive = false
			} else if tts.isBoardFull() {
				tts.gameBoard.GameStatus = 1
				tts.gameBoard.Winner = 3 // Draw
				tts.gameActive = false
			} else {
				// Switch turns
				if tts.gameBoard.CurrentTurn == protocol.X {
					tts.gameBoard.CurrentTurn = protocol.O
				} else {
					tts.gameBoard.CurrentTurn = protocol.X
				}
			}

			tts.broadcastGameState()
			tts.mutex.Unlock()
		}
	}

	// Clean up
	if currentPlayerID != 0 {
		tts.mutex.Lock()
		delete(tts.players, currentPlayerID)
		delete(tts.clients, currentPlayerID)
		tts.gameActive = false
		tts.gameBoard = protocol.GameBoard{CurrentTurn: protocol.X} // Reset
		tts.mutex.Unlock()

		log.Printf("Player %d disconnected", currentPlayerID)
	}
}

func (tts *TicTacToeServer) checkWinner() uint8 {
	board := tts.gameBoard.Board

	// Check rows
	for i := 0; i < 3; i++ {
		if board[i*3] != protocol.EMPTY &&
			board[i*3] == board[i*3+1] &&
			board[i*3] == board[i*3+2] {
			return board[i*3]
		}
	}

	// Check columns
	for i := 0; i < 3; i++ {
		if board[i] != protocol.EMPTY &&
			board[i] == board[i+3] &&
			board[i] == board[i+6] {
			return board[i]
		}
	}

	// Check diagonals
	if board[0] != protocol.EMPTY && board[0] == board[4] && board[0] == board[8] {
		return board[0]
	}
	if board[2] != protocol.EMPTY && board[2] == board[4] && board[2] == board[6] {
		return board[2]
	}

	return protocol.EMPTY
}

func (tts *TicTacToeServer) isBoardFull() bool {
	for _, cell := range tts.gameBoard.Board {
		if cell == protocol.EMPTY {
			return false
		}
	}
	return true
}

func (tts *TicTacToeServer) broadcastGameStart() {
	startMsg := protocol.EncodeTLV(protocol.MSG_GAME_START, []byte("Game started!"))
	for _, conn := range tts.clients {
		conn.Write(startMsg)
	}
	tts.broadcastGameState()
}

func (tts *TicTacToeServer) broadcastGameState() {
	boardData := protocol.EncodeGameBoard(tts.gameBoard)
	stateMsg := protocol.EncodeTLV(protocol.MSG_GAME_STATE, boardData)

	for _, conn := range tts.clients {
		conn.Write(stateMsg)
	}
}
