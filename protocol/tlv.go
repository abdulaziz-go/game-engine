package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Message types
const (
	MSG_PLAYER_JOIN  = 1
	MSG_MAKE_MOVE    = 2
	MSG_GAME_STATE   = 3
	MSG_GAME_OVER    = 4
	MSG_PLAYER_LEAVE = 5
	MSG_WAITING      = 6
	MSG_GAME_START   = 7
	MSG_ERROR        = 8
)

// Game constants
const (
	EMPTY = 0
	X     = 1
	O     = 2
)

// GameBoard represents the tic-tac-toe board
type GameBoard struct {
	Board       [9]uint8 // 3x3 board flattened
	CurrentTurn uint8    // X or O
	GameStatus  uint8    // 0=playing, 1=X wins, 2=O wins, 3=draw
	Winner      uint8    // 0=none, 1=X, 2=O, 3=draw
}

// Player represents a player
type Player struct {
	ID     uint32
	Name   string
	Symbol uint8 // X or O
}

// Move represents a game move
type Move struct {
	PlayerID uint32
	Position uint8 // 0-8 for board positions
}

// TLVMessage represents a Type-Length-Value message
type TLVMessage struct {
	Type   uint8
	Length uint16
	Value  []byte
}

// EncodeTLV encodes a message into TLV format
func EncodeTLV(msgType uint8, data []byte) []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.BigEndian, msgType)
	binary.Write(buf, binary.BigEndian, uint16(len(data)))
	buf.Write(data)

	return buf.Bytes()
}

// DecodeTLV decodes a TLV message
func DecodeTLV(data []byte) (*TLVMessage, error) {
	if len(data) < 3 {
		return nil, fmt.Errorf("insufficient data for TLV header")
	}

	msg := &TLVMessage{}
	buf := bytes.NewReader(data)

	binary.Read(buf, binary.BigEndian, &msg.Type)
	binary.Read(buf, binary.BigEndian, &msg.Length)

	if len(data) < int(3+msg.Length) {
		return nil, fmt.Errorf("insufficient data for TLV value")
	}

	msg.Value = data[3 : 3+msg.Length]
	return msg, nil
}

// EncodePlayer encodes a player to bytes
func EncodePlayer(player Player) []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.BigEndian, player.ID)
	binary.Write(buf, binary.BigEndian, player.Symbol)

	nameBytes := []byte(player.Name)
	binary.Write(buf, binary.BigEndian, uint8(len(nameBytes)))
	buf.Write(nameBytes)

	return buf.Bytes()
}

// DecodePlayer decodes bytes to player
func DecodePlayer(data []byte) (Player, error) {
	player := Player{}
	if len(data) < 6 {
		return player, fmt.Errorf("insufficient data")
	}

	buf := bytes.NewReader(data)
	binary.Read(buf, binary.BigEndian, &player.ID)
	binary.Read(buf, binary.BigEndian, &player.Symbol)

	var nameLen uint8
	binary.Read(buf, binary.BigEndian, &nameLen)

	if len(data) < int(6+nameLen) {
		return player, fmt.Errorf("insufficient data for name")
	}

	nameBytes := make([]byte, nameLen)
	buf.Read(nameBytes)
	player.Name = string(nameBytes)

	return player, nil
}

// EncodeGameBoard encodes game board to bytes
func EncodeGameBoard(board GameBoard) []byte {
	buf := new(bytes.Buffer)

	for _, cell := range board.Board {
		binary.Write(buf, binary.BigEndian, cell)
	}
	binary.Write(buf, binary.BigEndian, board.CurrentTurn)
	binary.Write(buf, binary.BigEndian, board.GameStatus)
	binary.Write(buf, binary.BigEndian, board.Winner)

	return buf.Bytes()
}

// DecodeGameBoard decodes bytes to game board
func DecodeGameBoard(data []byte) (GameBoard, error) {
	board := GameBoard{}
	if len(data) < 12 {
		return board, fmt.Errorf("insufficient data")
	}

	buf := bytes.NewReader(data)
	for i := 0; i < 9; i++ {
		binary.Read(buf, binary.BigEndian, &board.Board[i])
	}
	binary.Read(buf, binary.BigEndian, &board.CurrentTurn)
	binary.Read(buf, binary.BigEndian, &board.GameStatus)
	binary.Read(buf, binary.BigEndian, &board.Winner)

	return board, nil
}

// EncodeMove encodes a move to bytes
func EncodeMove(move Move) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, move.PlayerID)
	binary.Write(buf, binary.BigEndian, move.Position)
	return buf.Bytes()
}

// DecodeMove decodes bytes to move
func DecodeMove(data []byte) (Move, error) {
	move := Move{}
	if len(data) < 5 {
		return move, fmt.Errorf("insufficient data")
	}

	buf := bytes.NewReader(data)
	binary.Read(buf, binary.BigEndian, &move.PlayerID)
	binary.Read(buf, binary.BigEndian, &move.Position)

	return move, nil
}
