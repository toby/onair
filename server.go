package onair

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"os"
	"os/signal"
	"strings"
	"syscall"
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
	RegisterTrackOutChan(chan<- Track)
	Start()
}

// TrackSink provides an method for registering an output channel of played Tracks
type TrackSink interface {
	RegisterTrackInChan(<-chan Track)
}

// Server handles control messages and watches for shairport-sync metadata
type Server struct {
	port   int
	tracks chan Track
	source TrackSource
	sink   TrackSink
}

// NewServer returns a configured Server
func NewServer(port int, source TrackSource, sink TrackSink) Server {
	s := Server{
		port:   port,
		source: source,
		sink:   sink,
		tracks: make(chan Track, 0),
	}
	source.RegisterTrackOutChan(s.tracks)
	sink.RegisterTrackInChan(s.tracks)
	return s
}

// Tracks returns the output channel of Tracks from TrackSources
func (me *Server) Tracks() <-chan Track {
	return me.tracks
}

// Start watching for shairport-sync metadata
func (me *Server) Start() {
	address := net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: me.port}
	listener, err := net.ListenTCP("tcp", &address)
	defer listener.Close()
	if err != nil {
		if strings.Index(err.Error(), "in use") == -1 {
			panic(err)
		}
		fmt.Fprintln(os.Stderr, "Already running.")
		conn, err := net.DialTCP("tcp", &net.TCPAddr{IP: net.ParseIP("127.0.01")}, &address)
		defer conn.Close()
		if err != nil {
			log.Println("borked")
			panic(err)
		}
		log.Println("Connected")
		w := bufio.NewWriter(conn)
		tw := textproto.NewWriter(w)
		tw.PrintfLine("CMD %s", "skip")
	} else {
		go func() {
			for {
				conn, err := listener.AcceptTCP()
				if err != nil {
					if strings.Index(err.Error(), "closed network connection") == -1 {
						println("Error accept:", err.Error())
					}
					return
				}
				log.Printf("Connected: %v", conn)
				go me.handleConnection(conn)
			}
		}()
		me.source.Start()
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs
	}
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
