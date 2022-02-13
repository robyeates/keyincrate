package export

import (
	"github.com/tmthrgd/id3v2"
	"keyincrate/internal/seratosessions"
	"log"
)

func SessionToM3u8(session *seratosessions.Session, enrich bool) {
	log.Printf("Extracting m3u data from Session [%s] with optional enrichment [%v]", session.Id, enrich)
	for _, v := range session.LastPlayed {
		log.Printf("Extracting Frames from [%s]", v.TrackPath)
		frames, err := id3v2.ScanFile(v.TrackPath)
		if err != nil {
			log.Fatal(err)
		}
		title := getFrameText(frames.Lookup(id3v2.FrameTIT2))
		artist := getFrameText(frames.Lookup(id3v2.FrameTPE1))
		album := getFrameText(frames.Lookup(id3v2.FrameTALB))
		year := getFrameText(frames.Lookup(id3v2.FrameTYER))
		genre := getFrameText(frames.Lookup(id3v2.FrameTCON))
		comments := getFrameText(frames.Lookup(id3v2.FrameCOMM))
		bpm := getFrameText(frames.Lookup(id3v2.FrameTBPM))
		key := getFrameText(frames.Lookup(id3v2.FrameTKEY))
		size := getFrameText(frames.Lookup(id3v2.FrameTSIZ))
		length := getFrameText(frames.Lookup(id3v2.FrameTLEN))

		log.Printf(`Found from mp3.
					Title    [%s]
					Artist   [%s]
					Album    [%s]
					Year     [%s]
					Genre    [%s]
					Comments [%s]
					BPM      [%s]
					Key      [%s]
					Size     [%s]
					Length   [%s]`, title, artist, album, year, genre, comments, bpm, key, size, length)

		//TODO enrich from Serato DB

	}
}

func getFrameText(frame *id3v2.Frame) string {
	if frame != nil {
		frameText, err := frame.Text()
		if err != nil {
			log.Printf("Failed to get [%s] from [%v]", frame.ID, frameText)
		}
		return frameText
	}
	return ""
}
