package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/toby/onair"
)

func main() {
	p := flag.Int("p", 22212, "control port")
	v := flag.Bool("v", false, "verbose")
	s := flag.Bool("s", false, "echo blank newline when playback stops")
	a := flag.Bool("a", false, "display album name")
	m := flag.String("m", "/tmp/shairport-sync-metadata", "`path` to shairport-sync-metadata file")
	flag.Parse()
	if *v == false {
		log.SetOutput(ioutil.Discard)
	}
	log.Println("On Air")
	sc := onair.NewShairportClient(*m)
	so := onair.StdOut{ShowAlbum: *a, ShowPlaybackStop: *s}
	server := onair.NewServer(*p, &sc, &so)
	server.Start()
}
