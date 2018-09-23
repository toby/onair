# On Air

[![GoDoc](https://godoc.org/github.com/toby/onair?status.svg)](http://godoc.org/github.com/toby/onair)

`onair` manages music playback, metadata storage and display. It currently
supports shairport-sync as a playback source.

## Requirements

*  [shairport-sync](https://github.com/mikebrady/shairport-sync) (compiled `--with-metadata`)
*  Go 1.11+

## Installing

```
go get github.com/toby/onair/cmd/onair
```

## Usage

On Air will run continuously to print out metadata and act as a control for shairport-sync.

```
onair [flags] [COMMAND]
FLAGS:
  -h       Help
  -m PATH  Path to shairport-sync-metadata file (default "/tmp/shairport-sync-metadata")
  -a	   Display album name
  -s	   Print a blank newline when playback stops
  -v	   Verbose
COMMANDS
  skip     Skips to next track
  back     Play last track
  pause    Toggle pause
```

## Server Mode

Running `onair` with no commands prints each new track to standard out. If the
`-s` flag is supplied, `onair` will output a blank newline when there is a
stop in playback. This can be useful if you want to track your listening
sessions or for updating a UI to clear now playing information.

## Client Mode

When run with a command argument, `onair` will connect to an already running
`onair` server and tell it to issue the given command to the connected playback
device. If no server is running, you'll need to launch one. If the server has
not yet seen the required playback ids from the source, you may need to
reconnect your Airplay device.
