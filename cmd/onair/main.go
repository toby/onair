package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/toby/onair"
)

func main() {
	p := flag.Int("p", 22212, "control port")
	v := flag.Bool("v", false, "verbose")
	n := flag.Bool("n", false, "echo blank newline when playback stops")
	a := flag.Bool("a", false, "display album name")
	m := flag.String("m", "/tmp/shairport-sync-metadata", "`path` to shairport-sync-metadata file")
	flag.Parse()
	if *v == false {
		log.SetOutput(ioutil.Discard)
	}
	log.Println("On Air")
	cfg := onair.Config{}
	cfg.UseSessions = *n
	cfg.ShowAlbum = *a
	cfg.Port = *p
	s := onair.NewServer(cfg)
	sc := onair.NewShairportClient(*m)
	s.AddTrackSource(&sc)
	s.Start()
	for t := range s.Tracks() {
		if cfg.ShowAlbum {
			fmt.Printf("%s - %s - %s\n", t.Artist, t.Album, t.Name)
		} else {
			fmt.Printf("%s - %s\n", t.Artist, t.Name)
		}
	}
}
