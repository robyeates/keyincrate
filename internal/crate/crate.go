package crate

import (
	m "gopkg.in/music-theory.v0/key"
	"io/fs"
	"time"
)

type Crate struct {
	name         string
	partition    string
	fileInfo     fs.FileInfo
	lastModified time.Time
	trackData    []EnrichedCrateTrackData
}

type EnrichedCrateTrackData struct {
	trackName      string //join crate data to db - needs to be same partition
	shortTrackName string
	key            m.Key
	bpm            float32
	bitrateKbps    float32
	playCount      int
	genre          string
	duration       string
}
