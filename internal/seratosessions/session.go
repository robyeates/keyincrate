package seratosessions

import (
	"fmt"
	p "github.com/SpinTools/seratoparser"
	f "github.com/fsnotify/fsnotify"
	"io/fs"
	e "keyincrate/events"
	"keyincrate/internal/utils"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

var (
	seratoDirWithSession string
	sessionPath          = "/History/Sessions"
)

type Session struct {
	Id           string
	lastModified time.Time
	LastPlayed   map[int]SessionTrackPlayData
}

type SessionTrackPlayData struct {
	TrackName     string
	TrackPath     string
	StartDateTime string
	EndDateTime   string
	Deck          string
	Playtime      string
	LastModified  string
}

func SetSeratoDirectories(seratoDirs []string) <-chan int {
	for _, dir := range seratoDirs {
		maybeSessionDir := filepath.FromSlash(dir + sessionPath)
		file, err := os.Stat(maybeSessionDir)
		if err != nil {
			log.Printf("No Session Data found for [%s]", maybeSessionDir)
		} else {
			log.Printf("Session Data found for [%s] in [%s]", dir, file.Name())
			seratoDirWithSession = dir
			return make(chan int)
			//return watch(dir)
		}
	}
	return nil
}

func watch(sessionDir string) <-chan int {
	log.Printf("Creating File watcher for [%s]", sessionDir)
	watcher, err := f.NewWatcher()
	if err != nil {
		fmt.Println("Failed to create a new File Watcher", err)
	}
	defer func(watcher *f.Watcher) {
		err := watcher.Close()
		if err != nil {
			log.Printf("Failed to close file watcher for [%s]", sessionDir)
		}
	}(watcher)

	eventChannel := make(chan int)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&f.Write == f.Write {
					log.Printf("File watcher event received [%#v]", event)
					eventChannel <- e.SessionUpdated
				} else if event.Op&f.Create == f.Create {
					log.Printf("File watcher event received [%#v]", event)
					eventChannel <- e.SessionCreated
				} else {
					//just log it I guess
				}

			case err := <-watcher.Errors:
				log.Println("ERROR", err.Error())
			}
		}
	}()

	if err := watcher.Add(sessionDir); err != nil {
		log.Println("ERRORRRRRR", err)
	}

	return eventChannel
}

func GetSessionNames() []string {
	if seratoDirWithSession == "" {
		log.Print("No Serato Session Directories Available")
		return nil
	}
	log.Printf("Building Serato Sessions Rows for [%s]", seratoDirWithSession)
	parser := p.New(seratoDirWithSession)
	historySessions := parser.GetHistorySessions()
	log.Printf("Found [%v] Serato Sessions", len(historySessions))
	sessionNames := make([]string, len(historySessions))
	for i, v := range historySessions {
		sessionNames[i] = v.Name()
	}
	sort.Slice(sessionNames, func(p, q int) bool {
		return (sessionNames)[p] > (sessionNames)[q]
	})
	for _, v := range sessionNames {
		log.Printf(v)
	}
	return sessionNames
}

func GetSession(sessionFileName string) *Session {
	parser := p.New(seratoDirWithSession)
	sessions := parser.GetHistorySessions()
	historyEntries := parser.ReadHistorySession(sessionFileName)
	log.Printf("Found [%v] historyEntries for session [%s]", len(historyEntries), sessionFileName)
	if len(historyEntries) == 0 {
		log.Printf("No historyEntries found for session [%s]", sessionFileName)
	} else {
		return unpackHistoryEntries(sessions[0], historyEntries, len(historyEntries))
	}
	return nil
}

// GetLastFivePlayed decompose function
//last 5 (i.e whats playing)
//export session to m3u
func GetLastFivePlayed() *Session {
	if seratoDirWithSession == "" {
		log.Print("No Serato Session Directories Available")
		return nil
	}
	parser := p.New(seratoDirWithSession)
	sessions := parser.GetHistorySessions()
	if len(sessions) == 0 {
		log.Printf("No Session found for [%s]", seratoDirWithSession)
	} else {
		historyEntries := parser.ReadHistorySession(sessions[0].Name())
		if len(historyEntries) == 0 {
			log.Printf("No historyEntries found for session [%s]", sessions[0].Name())
		} else {
			return unpackHistoryEntries(sessions[0], historyEntries, 5)
		}
	}
	return nil
}

func unpackHistoryEntries(info fs.FileInfo, entries []p.HistoryEntity, i int) *Session {
	sort.Slice(entries, func(p, q int) bool {
		return entries[p].RROW > entries[q].RROW
	})
	//get the required number of values
	var sliced []p.HistoryEntity
	if len(entries) >= i {
		sliced = entries[0:i]
	} else {
		sliced = entries
	}
	playDataEntries := map[int]SessionTrackPlayData{}
	for _, entry := range sliced {
		playData := new(SessionTrackPlayData)
		playData.TrackName = filepath.Base(entry.RDIR)
		playData.TrackPath = entry.RDIR
		s, err := utils.EpochIntToTime(entry.TTMS)
		if err == nil {
			playData.StartDateTime = s.Format(time.Stamp)
		}
		eee, err := utils.EpochIntToTime(entry.TTME)
		if err == nil {
			playData.EndDateTime = eee.Format(time.Stamp)
		}
		playData.Deck = strconv.Itoa(entry.TDCK)
		playData.Playtime = strconv.Itoa(entry.TPTM)
		u, err := utils.EpochIntToTime(entry.RUPD)
		if err == nil {
			playData.LastModified = u.Format(time.Stamp)
		}
		if playData.TrackName != "" {
			log.Printf("Set Key [%v]", entry.RROW)
			playDataEntries[entry.RROW] = *playData
		}
	}
	session := new(Session)
	session.Id = filepath.Base(info.Name())
	session.lastModified = info.ModTime()
	session.LastPlayed = playDataEntries
	log.Printf("Created Session [%v]", session)
	return session
}
