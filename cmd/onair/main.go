package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/toby/onair"
)

const usage = `onair [flags] [COMMAND]

When no commands are issues, onair will run in server mode and watch for track
metadata. A server must be active and running to issue a command.

FLAGS:
  -h       Help
  -m PATH  Path to shairport-sync-metadata file (default "/tmp/shairport-sync-metadata")
  -p       onair control port (default: 22212)
  -u       shairport-sync metadata receive udp port (default: disabled)
  -a	   Display album name
  -s	   Print a blank newline when playback stops
  -v	   Verbose
COMMANDS
  display                       Displays currently playing track
  play                          Start playback
  pause                         Pause playback
  playpause                     Toggle between play and pause
  skip, next, nextitem          Play next item in playlist
  back, previous, previtem      Play previous item in playlist
  stop                          Stop playback
  shuffle, shuffle_songs        Shuffle playlist
  ff, fastforward, beginff      Begin fast forward
  rew, rewind, beginrew         Begin rewind
  playresume                    Play after fast forward or rewind
  up, volup, volumedown         Turn audio volume down
  down, voldown, volumeup       Turn audio volume up
  mute, mutetoggle              Toggle mute status
`

func defaultUsage() {
	fmt.Fprintf(flag.CommandLine.Output(), usage)
}

func main() {
	p := flag.Int("p", 22212, "onair control port")
	u := flag.Int("u", 0, "udp metadata port")
	v := flag.Bool("v", false, "verbose")
	s := flag.Bool("s", false, "echo blank newline when playback stops")
	a := flag.Bool("a", false, "display album name")
	m := flag.String("m", "/tmp/shairport-sync-metadata", "`path` to shairport-sync-metadata file")
	flag.Usage = defaultUsage
	flag.Parse()
	if *v == false {
		log.SetOutput(ioutil.Discard)
	}
	args := flag.Args()
	if len(args) == 0 { // No command sent, use server mode
		sp := onair.NewShairportClient(*m, *u)
		sink := onair.StdOut{ShowAlbum: *a, ShowPlaybackStop: *s}
		server := onair.NewServer(*p, &sp, &sink, &sp)
		server.Listen()
	} else { // Command supplied, use client mode
		client, err := onair.NewClient(*p)
		if err != nil {
			if strings.Index(err.Error(), "connection refused") != -1 {
				fmt.Println("Cannot connect to server, make sure `onair` is running with no commands")
			}
			log.Fatalln(err)
		}
		defer client.Close()
		err = client.Send(args[0])
		if err != nil {
			log.Fatalln(err)
		}
	}

}
