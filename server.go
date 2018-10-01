// Package onair provides a server and client for managing music playback,
// metadata storage and display. It currently supports shairport-sync as a
// playback source.
package onair

import (
	"bufio"
	"log"
	"net"
	"net/textproto"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// Track is the common model for an album track.
type Track struct {
	Artist   string
	Album    string
	Name     string
	Composer string
	Genre    string
	ID       uint64
	Time     uint32
}

// Server manages track flow and control commands.
type Server struct {
	port    int
	tracks  chan Track
	source  TrackSource
	sink    TrackSink
	control PlaybackControl
}

// TrackSource provides an interface playback sources need to implement. As
// tracks are played, they should be pushed into the provided channel. When
// playbacks stops, they should push a blank Track.
type TrackSource interface {
	// RegisterTrackOutChan supplies the source a Track output channel
	RegisterTrackOutChan(chan<- Track)
}

// TrackSink provides an interface track sinks need to implement. Sinks can
// pull tracks from the provided channel and store or render them as
// appropriate. When playback stops, sinks will receive a blank track.
type TrackSink interface {
	// RegisterTrackInChan supplies the sink a Track input channel
	RegisterTrackInChan(<-chan Track)
}

// PlaybackControl provides an interface for source specific playback
// controllers to implement.
type PlaybackControl interface {
	// Play starts playback.
	Play()
	// Pause pauses playback.
	Pause()
	// Next plays the next next item in the playlist.
	Next()
	// Previous plays the previous item in the playlist.
	Previous()
	// Stop playback.
	Stop()
	// FastForward begins fast forward, PlayResume() should be called to return to playback.
	FastForward()
	// Rewind begins rewinding, PlayResume() should be called to return to playback.
	Rewind()
	// PlayResume is called after a FastForward() or Rewind() call to resume playback.
	PlayResume()
	// TogglePause toggles pause state.
	TogglePause()
	// ToggleMute toggles mute state.
	ToggleMute()
	// Shuffle the tracks in a playlist.
	Shuffle()
	// VolumeUp increases the volume.
	VolumeUp()
	// VolumeDown decreases the volume.
	VolumeDown()
}

// NewServer returns a configured Server.
func NewServer(port int, source TrackSource, sink TrackSink, control PlaybackControl) Server {
	s := Server{
		port:    port,
		source:  source,
		sink:    sink,
		control: control,
		tracks:  make(chan Track, 0),
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
	if err != nil {
		if strings.Index(err.Error(), "in use") == -1 {
			panic(err)
		}
		log.Printf("Already listening")
		return
	}
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
}

func (me *Server) handleConnection(conn *net.TCPConn) {
	r := bufio.NewReader(conn)
	tp := textproto.NewReader(r)
	defer conn.Close()
	for {
		cmd, err := tp.ReadLine()
		if err != nil {
			break
		}
		switch cmd {
		case "play":
			me.control.Play()
		case "pause":
			me.control.Pause()
		case "nextitem":
			me.control.Next()
		case "previtem":
			me.control.Previous()
		case "stop":
			me.control.Stop()
		case "beginff":
			me.control.FastForward()
		case "beginrew":
			me.control.Rewind()
		case "playresume":
			me.control.PlayResume()
		case "playpause":
			me.control.TogglePause()
		case "mutetoggle":
			me.control.ToggleMute()
		case "shuffle_songs":
			me.control.Shuffle()
		case "volumedown":
			me.control.VolumeUp()
		case "volumeup":
			me.control.VolumeDown()
		default:
			log.Printf("Bad command: '%s'", cmd)
		}
	}
}
