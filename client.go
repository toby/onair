package onair

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/textproto"
)

// Commands are valid client commands for controlling playback. They mirror the
// Airport DACP commands.
var Commands = map[string]string{
	"play":          "Start playback",
	"pause":         "Pause playback",
	"playpause":     "Toggle between play and pause",
	"nextitem":      "Play next item in playlist",
	"previtem":      "Play previous item in playlist",
	"stop":          "Stop playback",
	"shuffle_songs": "Shuffle playlist",
	"beginff":       "Begin fast forward",
	"beginrew":      "Begin rewind",
	"playresume":    "Play after fast forward or rewind",
	"volumedown":    "Turn audio volume down",
	"volumeup":      "Turn audio volume up",
	"mutetoggle":    "Toggle mute status",
}

// CommandAliases offers more user friendly options for the DACP commands.
var CommandAliases = map[string]string{
	"next":        "nextitem",
	"skip":        "nextitem",
	"previous":    "previtem",
	"back":        "previtem",
	"shuffle":     "shuffle_songs",
	"fastforward": "beginff",
	"ff":          "beginff",
	"rewind":      "beginrew",
	"rew":         "beginrew",
	"up":          "volumeup",
	"volup":       "volumeup",
	"down":        "volumedown",
	"voldown":     "volumedown",
	"mute":        "mutetoggle",
}

// Client connects to an onair Server for sending playback control commands.
type Client struct {
	port   int
	writer *textproto.Writer
	conn   net.Conn
}

// NewClient attempts to create a new Client connected to a server port.
func NewClient(port int) (Client, error) {
	c := Client{
		port: port,
	}
	err := c.connect()
	return c, err
}

// Send will attempt to send a valid command to the server.
func (me *Client) Send(cmd string) error {
	alt, ok := CommandAliases[cmd]
	if ok {
		cmd = alt
	}
	_, ok = Commands[cmd]
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
