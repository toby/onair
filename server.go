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

// Server manages track flow and control commands
type Server struct {
	port   int
	tracks chan Track
	source TrackSource
	sink   TrackSink
}

// TrackSource allows a player to register a channel of Tracks to send when
// played.
type TrackSource interface {
	RegisterTrackOutChan(chan<- Track)
}

// TrackSink provides an method for registering a channel of played Tracks
type TrackSink interface {
	RegisterTrackInChan(<-chan Track)
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

// Listen for control commands. This method blocks until a termination signal
// is received.
func (me *Server) Listen() {
	address := net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: me.port}
	listener, err := net.ListenTCP("tcp", &address)
	defer listener.Close()
	if err == nil {
		log.Printf("Listening on port %d", me.port)
		go func() {
			for {
				conn, err := listener.AcceptTCP()
				if err != nil {
					if strings.Index(err.Error(), "closed network connection") == -1 {
						log.Printf("Error accept: %s", err.Error())
					}
					return
				}
				log.Printf("Connected: %v", conn)
				go me.handleConnection(conn)
			}
		}()
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs
	} else {
		if strings.Index(err.Error(), "in use") == -1 {
			panic(err)
		}
		log.Printf("Already listening")
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
