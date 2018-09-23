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
	s := flag.Bool("s", false, "echo blank newline when playback stops")
	a := flag.Bool("a", false, "display album name")
	m := flag.String("m", "/tmp/shairport-sync-metadata", "`path` to shairport-sync-metadata file")
	flag.Parse()
	if *v == false {
		log.SetOutput(ioutil.Discard)
	}
	args := flag.Args()
	if len(args) == 0 { // No command sent, use server mode
		sp := onair.NewShairportClient(*m)
		sink := onair.StdOut{ShowAlbum: *a, ShowPlaybackStop: *s}
		server := onair.NewServer(*p, &sp, &sink, &sp)
		server.Listen()
	} else { // Command supplied, use client mode
		client, err := onair.NewClient(*p)
		defer client.Close()
		if err != nil {
			panic(err)
		}
		err = client.Send(args[0])
		if err != nil {
			fmt.Println(err)
		}
	}

}
