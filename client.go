package onair

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/textproto"
)

// Client connects to an onair Server for sending playback control commands.
type Client struct {
	commands map[string]bool
	port     int
	writer   *textproto.Writer
	conn     net.Conn
}

// NewClient attempts to create a new Client connected to a server port.
func NewClient(port int) (Client, error) {
	commands := map[string]bool{
		"skip":  true,
		"back":  true,
		"pause": true,
	}
	c := Client{
		port:     port,
		commands: commands,
	}
	err := c.connect()
	return c, err
}

// Send will attempt to send a valid command to the server.
func (me *Client) Send(cmd string) error {
	_, ok := me.commands[cmd]
	if !ok {
		return fmt.Errorf("Invalid command: %s", cmd)
	}
	return me.writer.PrintfLine("%s", cmd)
}

// Close cleans up a client's network connections.
func (me *Client) Close() {
	me.conn.Close()
}

func (me *Client) connect() error {
	address := net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: me.port}
	conn, err := net.DialTCP("tcp", &net.TCPAddr{IP: net.ParseIP("127.0.01")}, &address)
	if err != nil {
		return err
	}
	log.Println("Connected")
	w := bufio.NewWriter(conn)
	me.conn = conn
	me.writer = textproto.NewWriter(w)
	return nil
}
