package client

import (
	"bufio"
	"fmt"
	"gamehomework/tlv"
	"net"
	"os"
	"strconv"
	"strings"
)

type Client struct {
	conn net.Conn
	name string
}

func Start() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Printf("error while connecting %v", err)
		return
	}
	defer conn.Close()

	client := &Client{conn: conn}
	client.run()
}

func (c *Client) run() {
	fmt.Printf("Enter your name: ")
	reader := bufio.NewReader(os.Stdin)
	name, _ := reader.ReadString('\n')
	c.name = strings.TrimSpace(name)

	c.conn.Write(tlv.Encode(tlv.JOIN, []byte(c.name)))

	go c.listenMessage()

	c.handleInput(reader)
}

func (c *Client) listenMessage() {
	for {
		buf := make([]byte, 1024)
		n, err := c.conn.Read(buf)

		if err != nil {
			return
		}
		msgType, data := tlv.Decode(buf[:n])

		switch msgType {
		case tlv.STATE:
			if len(data) >= 11 {
				c.displayBoard(data[:9], data[9], data[10])
			}
		case tlv.WIN:
			fmt.Println(string(data))
			return
		}
	}
}

func (c *Client) displayBoard(board []byte, turn, winner byte) {
	fmt.Print("\033[H\033[2J")

	symbols := map[byte]string{tlv.EMPTY: " ", tlv.X: "X", tlv.O: "O"}
	fmt.Printf("=== TIC-TAC-TOE === (Player: %s)\n", c.name)

	fmt.Println(" " + symbols[board[0]] + " | " + symbols[board[1]] + " | " + symbols[board[2]])
	fmt.Println("-----------")
	fmt.Println(" " + symbols[board[3]] + " | " + symbols[board[4]] + " | " + symbols[board[5]])
	fmt.Println("-----------")
	fmt.Println(" " + symbols[board[6]] + " | " + symbols[board[7]] + " | " + symbols[board[8]])

	if winner == 3 {
		fmt.Println("\n🤝 DRAW!")
	} else if winner != tlv.EMPTY {
		fmt.Printf("\n🎉 %s wins!\n", symbols[winner])
	} else {
		fmt.Printf("\nCurrent turn: %s\n", symbols[turn])
	}
}

func (c *Client) handleInput(reader *bufio.Reader) {
	fmt.Println("Connected! Enter moves as numbers 1-9:")
	fmt.Println("1 2 3")
	fmt.Println("4 5 6")
	fmt.Println("7 8 9")

	for {
		fmt.Print("Your move: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "q" {
			break
		}
		pos, err := strconv.Atoi(input)
		if err != nil || pos < 1 || pos > 9 {
			fmt.Println("Enter a number 1-9")
			continue
		}

		c.conn.Write(tlv.Encode(tlv.MOVE, []byte{byte(pos - 1)}))
	}
}
