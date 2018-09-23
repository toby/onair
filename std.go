package onair

import (
	"fmt"
)

// StdOut is a TrackSink that logs plays to the stdout optionally printing a
// newline when a playback stop event occurs.
type StdOut struct {
	ShowAlbum        bool
	ShowPlaybackStop bool
}

func (me *StdOut) printTrack(t Track) {
	if me.ShowAlbum {
		fmt.Printf("%s - %s - %s\n", t.Artist, t.Album, t.Name)
	} else {
		fmt.Printf("%s - %s\n", t.Artist, t.Name)
	}
}

// RegisterTrackInChan satisfies the TrackSink interface.
func (me *StdOut) RegisterTrackInChan(ts <-chan Track) {
	go func() {
		for t := range ts {
			blank := Track{}
			if t == blank {
				if me.ShowPlaybackStop {
					fmt.Println()
				}
				continue
			}
			me.printTrack(t)
		}
	}()
}
