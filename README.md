# On Air

[![GoDoc](https://godoc.org/github.com/toby/onair?status.svg)](http://godoc.org/github.com/toby/onair)

On Air manages music playback, metadata storage and display. It currently
supports shairport-sync as a playback source.

## Requirements

*  [shairport-sync](https://github.com/mikebrady/shairport-sync) (compiled `--with-metadata`)
*  Go 1.11+

## Installing

```
go get github.com/toby/onair/cmd/onair
```

## Usage

After starting, `onair` will display `ARTIST - ALBUM - TRACK` on new lines as
each track plays.

```
onair [flags] [COMMAND]
FLAGS:
  -h       Help
  -m PATH  Path to shairport-sync-metadata file (default "/tmp/shairport-sync-metadata")
  -a	   Display album name
  -n	   Print a blank newline when playback stops
  -v	   Verbose
COMMANDS
  skip     Skips to next track
  back     Play last track
  pause    Toggle pause
```
