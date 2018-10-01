package onair

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/grandcat/zeroconf"
)

// ShairportClient watches the shairport-sync-metadata file and handles Track parsing.
type ShairportClient struct {
	tracks       chan<- Track
	lastID       uint64
	track        Track
	dacpID       string
	remoteToken  string
	remotePort   string
	remoteHost   net.IP
	metadataPath string
}

// ServiceName returns the DACP mDNS service name based on the dacpID.
func (me *ShairportClient) ServiceName() string {
	return fmt.Sprintf("iTunes_Ctrl_%s", me.dacpID)
}

// MetadataItem is an XML entry from the shairport-sync-metadata file.
type MetadataItem struct {
	Type        string `xml:"type"`
	Code        string `xml:"code"`
	Length      int    `xml:"length"`
	EncodedData []byte `xml:"data"`
}

// NewShairportClient returns a ShairportClient that watches metadataPath for shairport-sync
// metadata.
func NewShairportClient(metadataPath string) ShairportClient {
	return ShairportClient{metadataPath: metadataPath}
}

// RegisterTrackOutChan satisfied the onair.TrackSource interface.
func (me *ShairportClient) RegisterTrackOutChan(c chan<- Track) {
	me.tracks = c
	me.start()
}

// Data decodes the base64 data stored in an item.
func (me *MetadataItem) Data() []byte {
	d := make([]byte, base64.StdEncoding.DecodedLen(len(me.EncodedData)))
	_, err := base64.StdEncoding.Decode(d, me.EncodedData)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return d[:me.Length]
}

// Play starts playback.
func (me *ShairportClient) Play() {
	me.clientRequest("play")
}

// Pause pauses playback.
func (me *ShairportClient) Pause() {
	me.clientRequest("pause")
}

// Next plays the next next item in the playlist.
func (me *ShairportClient) Next() {
	me.clientRequest("nextitem")
}

// Previous plays the previous item in the playlist.
func (me *ShairportClient) Previous() {
	me.clientRequest("previtem")
}

// Stop playback.
func (me *ShairportClient) Stop() {
	me.clientRequest("stop")
}

// FastForward begins fast forward, PlayResume() should be called to return to playback.
func (me *ShairportClient) FastForward() {
	me.clientRequest("beginff")
}

// Rewind begins rewinding, PlayResume() should be called to return to playback.
func (me *ShairportClient) Rewind() {
	me.clientRequest("beginrew")
}

// PlayResume is called after a FastForward() or Rewind() call to resume playback.
func (me *ShairportClient) PlayResume() {
	me.clientRequest("playresume")
}

// TogglePause toggles pause state.
func (me *ShairportClient) TogglePause() {
	me.clientRequest("playpause")
}

// ToggleMute toggles mute state.
func (me *ShairportClient) ToggleMute() {
	me.clientRequest("mutetoggle")
}

// Shuffle the tracks in a playlist.
func (me *ShairportClient) Shuffle() {
	me.clientRequest("shuffle_songs")
}

// VolumeUp increases the volume.
func (me *ShairportClient) VolumeUp() {
	me.clientRequest("volumeup")
}

// VolumeDown decreases the volume.
func (me *ShairportClient) VolumeDown() {
	me.clientRequest("volumedown")
}

// Start watching for shairport-sync metadata.
func (me *ShairportClient) start() {
	go func() {
		f, err := os.Open(me.metadataPath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer f.Close()
		decoder := xml.NewDecoder(f)
		for {
			t, _ := decoder.Token()
			if t == nil {
				log.Println("No more XML")
				break
			}
			switch v := t.(type) {
			case xml.StartElement:
				var i MetadataItem
				decoder.DecodeElement(&i, &v)
				err := i.decode()
				if err == nil {
					me.handleMetadataItem(&i)
				} else {
					log.Printf("Invalid item: %s\n", err)
				}
			}
		}
	}()
}

func (me *ShairportClient) connectCtrlService() error {
	resolver, err := zeroconf.NewResolver()
	if err != nil {
		log.Fatalln("Failed to initialize resolver:", err.Error())
	}

	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			if strings.Index(entry.Instance, me.dacpID) != -1 {
				log.Printf("mDNS Instance %s\n", entry.Instance)
				log.Printf("mDNS Port %d\n", entry.Port)
				if len(entry.AddrIPv4) > 0 {
					log.Println("Found matching service record")
					me.remoteHost = entry.AddrIPv4[0]
				}
				return
			}
		}
		log.Println("No more entries.")
	}(entries)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err = resolver.Browse(ctx, "_dacp._tcp", "local.", entries)
	if err != nil {
		log.Fatalln("Failed to browse:", err.Error())
	}

	<-ctx.Done()
	return nil
}

func (me *ShairportClient) clientRequest(cmd string) {
	if me.remoteHost == nil {
		log.Printf("Cannot send iTunes command: %s, have not received airport connect messages yet. Try reconnecting your iTunes source", cmd)
		return
	}
	client := &http.Client{}
	url := fmt.Sprintf("http://%s:%s/ctrl-int/1/%s", me.remoteHost.String(), me.remotePort, cmd)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Add("Active-Remote", me.remoteToken)
	log.Printf("iTunes request: %s", url)
	res, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Request status: %s", res.Status)
}

func (me *ShairportClient) handleMetadataItem(i *MetadataItem) {
	switch i.Code {
	case "pbeg":
		log.Println("Play stream begin")
	case "pend":
		log.Println("Play stream end")
		me.lastID = 0
		me.tracks <- Track{}
	case "pfls":
		log.Println("Play stream flush")
	case "prsm":
		log.Println("Play stream resume")
	case "mdst":
		log.Println("Metadata start")
		me.track = Track{}
	case "mden":
		log.Println("Metadata end")
		if me.lastID != me.track.ID {
			me.lastID = me.track.ID
			me.tracks <- me.track
		}
	case "asal":
		a := string(i.Data())
		log.Printf("Album:\t\t%s\n", a)
		me.track.Album = a
	case "asar":
		a := string(i.Data())
		log.Printf("Artist:\t\t%s\n", a)
		me.track.Artist = a
	case "ascp":
		c := string(i.Data())
		log.Printf("Composer:\t%s\n", c)
		me.track.Composer = c
	case "astm":
		t, err := byteUInt32(i.Data())
		if err != nil {
			log.Printf("bad astm: %s\n", err)
		}
		log.Printf("Time:\t\t%d\n", t)
		me.track.Time = t
	case "asgn":
		g := string(i.Data())
		log.Printf("Genre:\t\t%s\n", g)
		me.track.Genre = g
	case "minm":
		n := string(i.Data())
		log.Printf("Name:\t\t%s\n", n)
		me.track.Name = n
	case "caps":
		s, err := byteUInt8(i.Data())
		if err != nil {
			log.Printf("bad caps: %s\n", err)
		}
		log.Printf("Play Status:\t%v\n", s)
	case "mper":
		id, err := byteUInt64(i.Data())
		if err != nil {
			log.Printf("bad mper: %s\n", err)
		}
		log.Printf("ID\t\t%d\n", id)
		me.track.ID = id
	case "daid":
		d := string(i.Data())
		log.Printf("DACP-ID:\t\t%s\n", d)
		me.dacpID = d
		go me.connectCtrlService()
	case "acre":
		t := string(i.Data())
		log.Printf("Active-Remote Token:\t\t%s\n", t)
		me.remoteToken = t
	case "dapo":
		p := string(i.Data())
		log.Printf("Control port:\t\t%s\n", p)
		me.remotePort = p
	default:
		log.Printf("Unlogged:\t%s %s\n", i.Type, i.Code)
	}
}

func (me *MetadataItem) decode() error {
	t, err := hex.DecodeString(me.Type)
	if err != nil {
		return err
	}
	me.Type = string(t)
	c, err := hex.DecodeString(me.Code)
	if err != nil {
		return err
	}
	me.Code = string(c)
	return nil
}
