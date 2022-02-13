package ui

import (
	"embed"
	_ "embed"
	"fmt"
	g "github.com/AllenDang/giu"
	"image"
	"image/color"
	"image/png"
	_ "image/png"
	"keyincrate/internal/export"
	"keyincrate/internal/seratodatabase"
	"keyincrate/internal/seratosessions"
	"log"
	"os"
	"sort"
)

var (
	//config
	//bigFont *g.FontInfo
	enrichM3uFromSeratoDB = false
	logoImage             image.Image
	iconImage             image.Image
	//updateables
	deckOneSession          *g.LabelWidget
	deckTwoSession          *g.LabelWidget
	sessionNames            []string
	seratoFolders           []string
	foundSeratoFolders      []string
	sessionSessionName      []string
	selectedSessionFileName string
	sessionTableContents    []*g.TableRowWidget

	//chans
	partitionChan   chan string
	foundFolderChan chan string
	progressChan    chan float32

	//receiving vars
	searchPartition     string
	initialLoadProgress float32

	//state
	sync         = false
	itemSelected int32
	placeholder  []*g.TableRowWidget

	//channels
	sessionUpdateChannel <-chan int
)

//go:embed key_in_crate.png crate.png
var emb embed.FS //TODO mention in PR for image
//<a href="https://www.flaticon.com/free-icons/crate" title="crate icons">Crate icons created by Smashicons - Flaticon</a>

var initd = true

func loop() {
	if initd {
		w, h := g.GetAvailableRegion()
		select {
		case s := <-partitionChan:
			searchPartition = s
		case f := <-progressChan:
			if f > 0 {
				initialLoadProgress = f
			}
		case ff := <-foundFolderChan:
			if ff != "" {
				foundSeratoFolders = append(foundSeratoFolders, ff)
			}
		}
		g.Window("Loading").Pos(w/2, h/2).Size(810, 500).Layout(
			g.Align(g.AlignCenter).To(
				g.ImageWithRgba(logoImage).OnClick(func() {
					fmt.Println("image from file was clicked")
				}).Size(622, 109),
			),
			g.ProgressBar(initialLoadProgress).Size(g.Auto, 20),
			g.Dummy(5, 20),
			g.Condition(initialLoadProgress != 0.999, g.Layout{
				g.Align(g.AlignCenter).To(
					g.Label("Looking for Serato Folders in " + searchPartition + ""),
				),
			}, g.Layout{
				g.Dummy(5, 20),
			}),
			g.Align(g.AlignCenter).To(
				g.Table().Size(411, 150).Rows(getFolderRows(foundSeratoFolders)...).
					Columns(g.TableColumn("Serato Folder Paths")).
					Flags(g.TableFlagsHideable),
			),
			g.Condition(initialLoadProgress == 0.999, g.Layout{
				//window with found paths in table

				//enter text for extra with validation
				//ok
				g.Align(g.AlignCenter).To(g.Button("Looks Good").OnClick(func() {
					initd = false
				})),
			}, g.Layout{
				g.Dummy(5, 20),
			}),

			g.Align(g.AlignCenter).To(g.Button("Exit").OnClick(func() {
				os.Exit(0)
			})),
		)
		g.Update()
	} else {
		g.SingleWindowWithMenuBar().Layout(

			g.SplitLayout(g.DirectionVertical, 100,
				g.Layout{
					g.SplitLayout(g.DirectionHorizontal, 580,
						g.Layout{
							g.Align(g.AlignLeft).To(
								g.Row(
									g.BulletTextf("1"),
									g.Dummy(10, 0),
									deckOneSession,
								),
								g.Row(),
							),
						},
						g.Layout{
							g.Align(g.AlignRight).To(
								g.Row(
									deckTwoSession,
									g.Dummy(10, 0),
									g.BulletTextf("2"),
									g.Dummy(10, 0),
								)),
							g.Align(g.AlignRight).To(
								g.Row(
									g.Checkbox("Sync with External", &sync),
									g.Dummy(10, 0),
								),
							),
						}),
				},
				g.SplitLayout(g.DirectionVertical, 600,
					//main section
					g.Layout{},
					g.SplitLayout(g.DirectionHorizontal, 590,
						//session
						g.Layout{
							g.Row(
								g.Combo("", sessionNames[itemSelected], sessionNames, &itemSelected).OnChange(sessionFileChanged),

								//g.Dummy(20, 0),
								g.Button("Export to m3u").OnClick(exportToM3u),
								g.Tooltip("Export Session Track Data to an m3u8 format file. This can be used by any audio tool"),
								g.Checkbox("Enrich", &enrichM3uFromSeratoDB),
								g.Tooltip("If fields are missing from the original audio files, use the Serato DB analysis to populate them"),
							),
							g.Table().Freeze(0, 1).Columns(
								g.TableColumn("TrackName").InnerWidthOrWeight(5),
								g.TableColumn("Deck"),
								g.TableColumn("Playtime"),
								g.TableColumn("LastModified"),
							).Rows(sessionTableContents...).Flags(g.TableFlagsSizingStretchSame), //no flag for sizing will kick a runtime error
						},
						//crates
						g.Layout{},
					),
				),
			),
		)
	}
}

func getFolderRows(folders []string) []*g.TableRowWidget {
	rows := make([]*g.TableRowWidget, len(folders))
	for i, v := range folders {
		rows[i] = g.TableRow(g.Label(v))
	}
	log.Printf("Made Folder Rows [%v]", folders)
	if len(rows) == 0 {
		return placeholder
	}
	return rows
}

func exportToM3u() {
	//TODO Progress Bar
	//TODO GOROUTINES AND WAIT, is SSSLLLOOWWW
	log.Print("Building Serato Sessions m3u")
	if selectedSessionFileName == "" {
		selectedSessionFileName = sessionNames[0]
	}
	session := seratosessions.GetSession(selectedSessionFileName)
	log.Printf("Exporting Session [%s] with Enrichment[%v]", selectedSessionFileName, enrichM3uFromSeratoDB)
	export.SessionToM3u8(session, enrichM3uFromSeratoDB)
}

func sessionFileChanged() {
	log.Printf("Selected Serato Sessions [%s]", sessionNames[itemSelected])
	selectedSessionFileName = sessionNames[itemSelected]
	sessionTableContents = buildSessionRows()
	//update sessionTableContents
	g.Update()
}

func buildSessionRows() []*g.TableRowWidget {
	log.Printf("Building Serato Sessions Rows for Session [%s]", selectedSessionFileName)
	if selectedSessionFileName == "" {
		selectedSessionFileName = sessionNames[0]
	}
	session := seratosessions.GetSession(selectedSessionFileName)
	rows := make([]*g.TableRowWidget, len(session.LastPlayed))

	for i, v := range getKeys(session.LastPlayed) {
		log.Printf("Map Key [%v]", v)
		rows[i] = g.TableRow(
			g.Label(session.LastPlayed[v].TrackName),
			g.Label(session.LastPlayed[v].Deck),
			g.Label(session.LastPlayed[v].Playtime),
			g.Label(session.LastPlayed[v].LastModified),
		)
		log.Printf("Row [%s], [%s], [%s], [%s]", session.LastPlayed[i].TrackName,
			session.LastPlayed[v].Deck,
			session.LastPlayed[v].Playtime,
			session.LastPlayed[v].LastModified)
	}

	rows[0].BgColor(&(color.RGBA{R: 200, G: 100, B: 100, A: 255}))
	log.Printf("Building Serato Sessions Rows - Complete with [%v] Rows [%v]", len(rows), rows)
	return rows
}

func getKeys(m map[int]seratosessions.SessionTrackPlayData) []int {
	//The default length of the array is the length of the map. When the array is attached, there is no need to re apply for memory and copy, which is very efficient
	j := 0
	keys := make([]int, len(m))
	for k := range m {
		keys[j] = k
		j++
	}
	log.Printf("Returnings Keys [%v]", keys)
	sort.Ints(keys)
	return keys
}

func InitState(part chan string, pro chan float32, found chan string) {
	partitionChan <- ""
	progressChan <- 0.0
	found <- ""

	placeholder = make([]*g.TableRowWidget, 1)
	placeholder[0] = g.TableRow(g.Label(""))
	deckOneSession = g.Label("Nothing played on Deck 1")
	deckTwoSession = g.Label("Nothing played on Deck 2")

	seratoFolders = seratodatabase.GetAllSeratoFolders(part, pro, found)

	seratosessions.SetSeratoDirectories(seratoFolders)

	sessionNames = seratosessions.GetSessionNames()
	log.Printf("Set [%v] Serato Sessions", len(sessionNames))
}

func InitUi() {
	partitionChan = make(chan string)
	progressChan = make(chan float32)
	foundFolderChan = make(chan string)

	initLogs()
	go InitState(partitionChan, progressChan, foundFolderChan)

	wnd := g.NewMasterWindow("transparent", 1200, 1000, g.MasterWindowFlagsNotResizable|g.MasterWindowFlagsTransparent|g.MasterWindowFlagsFrameless)
	wnd.SetBgColor(color.RGBA{43, 43, 43, 200})
	hmm := []image.Image{iconImage}
	wnd.SetIcon(hmm)
	wnd.Run(loop)
}

func initLogs() {
	f, err := os.OpenFile("golang-demo.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	defer f.Close()
	log.SetOutput(f)
	file, err := emb.Open("key_in_crate.png")
	icon, err := emb.Open("crate.png")
	stat, err := file.Stat()
	if err != nil {
		log.Printf(err.Error())
	}
	log.Printf("Image exists [%v]", stat.Name())

	logoImage, err = png.Decode(file)
	iconImage, err = png.Decode(icon)
	//bigFont = g.AddFont("Menlo.ttc", 24)
}
