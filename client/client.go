package client

import (
	"bufio"
	"fmt"
	"gamehomework/tlv"
	"net"
	"os"
)

type Client struct {
	conn net.Conn
	name string
}

func NewClient() *Client {
	return &Client{}
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
	c.name = name

	c.conn.Write(tlv.Encode(tlv.JOIN, []byte(c.name)))

	go c.listenMessage()
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
			c.displayBoard(data)
		case tlv.WIN:
			fmt.Println(string(data))
			return
		}
	}
}
