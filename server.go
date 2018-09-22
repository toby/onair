package onair

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"os"
	"strings"
)

// Track is the common model for an album track
type Track struct {
	Artist   string
	Album    string
	Name     string
	Composer string
	Genre    string
	ID       uint64
	Time     uint32
}

// TrackSource allows a player to register a channel of Tracks to send when
// played.
type TrackSource interface {
	RegisterTrackChan(chan<- Track)
	Start()
}

// Server handles control messages and watches for shairport-sync metadata
type Server struct {
	tracks chan Track
	config Config
	source TrackSource
}

// Config contains needed settings for Server
type Config struct {
	MetadataPath string
	UseSessions  bool
	ShowAlbum    bool
	Port         int
}

// NewServer returns a configured Server
func NewServer(cfg Config) Server {
	s := Server{
		config: cfg,
		tracks: make(chan Track, 0),
	}
	return s
}

// Tracks returns the output channel of Tracks from TrackSources
func (me *Server) Tracks() <-chan Track {
	return me.tracks
}

// AddTrackSource registers a TrackSource plugin that will supply tracks
func (me *Server) AddTrackSource(ts TrackSource) {
	ts.RegisterTrackChan(me.tracks)
	me.source = ts
}

// Start watching for shairport-sync metadata
func (me *Server) Start() {
	address := net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: me.config.Port}
	listener, err := net.ListenTCP("tcp", &address)
	if err != nil {
		if strings.Index(err.Error(), "in use") == -1 {
			panic(err)
		}
		fmt.Fprintln(os.Stderr, "Already running.")
		conn, err := net.DialTCP("tcp", &net.TCPAddr{IP: net.ParseIP("127.0.01")}, &address)
		if err != nil {
			log.Println("borked")
			panic(err)
		}
		log.Println("Connected")
		w := bufio.NewWriter(conn)
		tw := textproto.NewWriter(w)
		tw.PrintfLine("CMD %s", "skip")
		return
	}
	go func() {
		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				println("Error accept:", err.Error())
				return
			}
			log.Printf("Connected: %v", conn)
			go me.handleConnection(conn)
		}
	}()
	me.source.Start()
}

func (me *Server) handleConnection(conn *net.TCPConn) {
	r := bufio.NewReader(conn)
	tp := textproto.NewReader(r)
	defer conn.Close()
	for {
		line, err := tp.ReadLine()
		if err != nil {
			break
		}
		fmt.Printf("%s\n", line)
	}
}
