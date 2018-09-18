# On Air

A simple client for displaying track metadata from shairport-sync.

## Requirements

*  [shairport-sync](https://github.com/mikebrady/shairport-sync) (compiled `--with-metadata`)
*  Go 1.11+

## Installing

```
go get github.com/toby/onair
```

## Usage

After starting, `onair` will display `ARTIST - ALBUM - TRACK` on new lines as
each track plays.

```
onair [flags]
FLAGS:
  -h       Help
  -m PATH  Path to shairport-sync-metadata file (default "/tmp/shairport-sync-metadata")
  -v	   Verbose
```
