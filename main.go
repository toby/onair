package main

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

// Track is generated from a sequence of metadata items from
// shairport-sync-metadata.
type Track struct {
	Name     string
	Artist   string
	Album    string
	Composer string
	Genre    string
	ID       uint64
	Time     uint32
}

// Client watches the shairport-sync-metadata file and handles Track parsing.
type Client struct {
	metadataPath string
	lastID       uint64
	track        Track
	sessions     bool
	showAlbum    bool
	dacpID       string
	remoteToken  string
	remotePort   string
}

// Item is an XML entry from the shairport-sync-metadata file.
type Item struct {
	Type        string `xml:"type"`
	Code        string `xml:"code"`
	Length      int    `xml:"length"`
	EncodedData []byte `xml:"data"`
}

// Data decodes the base64 data stored in an item.
func (me *Item) Data() []byte {
	d := make([]byte, base64.StdEncoding.DecodedLen(len(me.EncodedData)))
	_, err := base64.StdEncoding.Decode(d, me.EncodedData)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return d[:me.Length]
}

func (me *Item) decode() error {
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

func (me *Client) handle(i *Item) {
	switch i.Code {
	case "pbeg":
		log.Println("Play stream begin")
	case "pend":
		log.Println("Play stream end")
		if me.sessions {
			fmt.Println()
			me.lastID = 0
		}
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
			if me.showAlbum {
				fmt.Printf("%s - %s - %s\n", me.track.Artist, me.track.Album, me.track.Name)
			} else {
				fmt.Printf("%s - %s\n", me.track.Artist, me.track.Name)
			}
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

func (me *Client) open() {
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
			var i Item
			decoder.DecodeElement(&i, &v)
			err := i.decode()
			if err == nil {
				me.handle(&i)
			} else {
				log.Printf("Invalid item: %s\n", err)
			}
		}
	}
}

func main() {
	v := flag.Bool("v", false, "verbose")
	n := flag.Bool("n", false, "echo blank newline when playback stops")
	a := flag.Bool("a", false, "display album name")
	m := flag.String("m", "/tmp/shairport-sync-metadata", "`path` to shairport-sync-metadata file")
	flag.Parse()
	if *v == false {
		log.SetOutput(ioutil.Discard)
	}
	log.Println("On Air")
	c := Client{}
	c.metadataPath = *m
	c.sessions = *n
	c.showAlbum = *a
	c.open()
}
