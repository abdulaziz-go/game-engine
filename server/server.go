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
	} else {
		room = &GameRoom{
			id:    s.roomId,
			board: game.NewBoard(),
		}
	}

	s.rooms[s.roomId] = room
	s.waitingRoom = room
	s.roomId++

	room.mutex.Lock()
	playerNum := room.count
	room.players[playerNum] = conn
	if playerNum == 1 {
		room.symbols[playerNum] = tlv.O
	}

	room.count++

	fmt.Printf("Player %d joined room #%d as %s\n", playerNum+1, room.id,
		map[byte]string{tlv.X: "X", tlv.O: "O"}[room.symbols[playerNum]])

	if room.count == 2 {
		if s.waitingRoom == room {
			s.waitingRoom = nil
		}

		fmt.Printf("Room is full game started %d", room.id)
	}

	room.mutex.Unlock()
	s.mutex.Unlock()

	room.handlePlayer(conn, playerNum)
}

func (room *GameRoom) handlePlayer(conn net.Conn, playerNum int) {
	defer conn.Close()

	room.mutex.Lock()
	defer room.mutex.Unlock()

	msg := tlv.Encode(tlv.STATE, room.board.ToBytes())

	for i := range room.count {
		if room.players[i] != nil {
			room.players[i].Write(msg)
		}
	}

	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			break
		}

		msgType, data := tlv.Decode(buf[:n])
		if msgType != tlv.MOVE || len(data) < 1 {
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
