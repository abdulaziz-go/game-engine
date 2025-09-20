package server

import (
	"fmt"
	"gamehomework/game"
	"gamehomework/tlv"
	"log"
	"net"
	"sync"
)

type Server struct {
	rooms       map[int]*GameRoom
	waitingRoom *GameRoom
	roomId      int
	mutex       sync.Mutex
}

type GameRoom struct {
	id      int
	board   *game.Board
	players [2]net.Conn
	symbols [2]byte
	count   int
	mutex   sync.Mutex
}

func NewServer() *Server {
	return &Server{
		rooms:  make(map[int]*GameRoom),
		roomId: 1,
	}
}

func Start() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	fmt.Println("Tic-Tac-Toe server running on :8080")

	server := NewServer()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go server.handleNewClient(conn)
	}
}

func (s *Server) handleNewClient(conn net.Conn) {
	s.mutex.Lock()
	var room *GameRoom
	if s.waitingRoom != nil && s.waitingRoom.count < 2 {
		room = s.waitingRoom
		fmt.Printf("Player joining existing room #%d\n", room.id)
	} else {
		room = &GameRoom{
			id:    s.roomId,
			board: game.NewBoard(),
		}
		s.rooms[s.roomId] = room
		s.waitingRoom = room
		fmt.Printf("Created new game room #%d\n", room.id)
		s.roomId++
	}

	room.mutex.Lock()
	playerNum := room.count
	room.players[playerNum] = conn

	room.symbols[0] = tlv.X
	room.symbols[1] = tlv.O

	room.count++

	fmt.Printf("Player %d joined room #%d as %s\n", playerNum+1, room.id,
		map[byte]string{tlv.X: "X", tlv.O: "O"}[room.symbols[playerNum]])

	if room.count == 2 {
		if s.waitingRoom == room {
			s.waitingRoom = nil
		}
		fmt.Printf("Room #%d is full - game started!\n", room.id)
	}

	room.mutex.Unlock()
	s.mutex.Unlock()

	room.handlePlayer(conn, playerNum)
}

func (room *GameRoom) handlePlayer(conn net.Conn, playerNum int) {
	defer conn.Close()

	room.sendState()

	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			break
		}
		msgType, data := tlv.Decode(buf[:n])
		if msgType != tlv.MOVE || len(data) < 1 {
			fmt.Println("error while decoding")
			continue
		}

		pos := int(data[0])
		room.mutex.Lock()
		symbol := room.symbols[playerNum]
		success := room.board.MakeMove(pos, symbol)
		room.mutex.Unlock()

		if success {
			room.sendState()
		}
	}

	room.handleDisconnect(playerNum)
}

func (room *GameRoom) sendState() {
	room.mutex.Lock()
	defer room.mutex.Unlock()

	data := room.board.ToBytes()
	msg := tlv.Encode(tlv.STATE, data)

	for i := 0; i < room.count; i++ {
		if room.players[i] != nil {
			room.players[i].Write(msg)
		}
	}
}

func (room *GameRoom) handleDisconnect(playerNum int) {
	room.mutex.Lock()
	defer room.mutex.Unlock()

	fmt.Printf("Player %d disconnected from room #%d\n", playerNum+1, room.id)
	room.players[playerNum] = nil

	disconnectMsg := tlv.Encode(tlv.WIN, []byte("Opponent disconnected! Game ended."))
	for i := 0; i < room.count; i++ {
		if room.players[i] != nil {
			room.players[i].Write(disconnectMsg)
			room.players[i].Close()
			room.players[i] = nil
		}
	}

	fmt.Printf("Room #%d closed\n", room.id)
}
