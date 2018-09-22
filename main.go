package main

import (
	"flag"
	// "fmt"
	"io/ioutil"
	"log"
	// "net"
	// "net/textproto"
	// "os"
	// "strings"

	"github.com/toby/onair/shairport"
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
	c := shairport.Client{}
	cfg := shairport.Config{}
	cfg.MetadataPath = *m
	cfg.Sessions = *n
	cfg.ShowAlbum = *a
	cfg.Port = *p
	c.Config = cfg
	c.Start()
}
