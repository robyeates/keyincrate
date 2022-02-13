package seratodatabase

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
	"time"

	d "github.com/shirou/gopsutil/v3/disk"
)

func GetAllSeratoFolders(part chan string, pro chan float32, found chan string) []string {
	start := time.Now()
	partitions, err := d.Partitions(true)
	pro <- 0.1

	currentUser, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	userHomeDir := filepath.FromSlash(currentUser.HomeDir)

	var startPoints []string
	var seratoFolders []string
	startPoints = append(startPoints, userHomeDir)

	for _, v := range partitions {
		startPoints = append(startPoints, v.Mountpoint)
	}
	search := float32(0.2)
	pro <- search
	//barTemplate := `{{ green "Partitions:" }} {{counters . "%s/%s" }} {{ bar . "[" "░" (cycle . "█" "▀" "█" "▄" ) "." "]"}} {{percent . "%.0f%%" | green}}`
	//bar := pb.ProgressBarTemplate(barTemplate).Start(len(startPoints))

	incr := 0.8 / float32(len(startPoints))

	for _, startPoint := range startPoints {
		part <- startPoint
		tmp := findSeratoFolder(startPoint)
		seratoFolders = append(seratoFolders, tmp...)
		for _, ff := range tmp {
			found <- ff
		}
		search += incr
		pro <- search
	}
	pro <- 0.999
	part <- ""
	//bar.Finish()
	for _, seratoFolder := range seratoFolders {
		stat, err := os.Stat(seratoFolder)
		if err != nil {
			return nil
		}
		log.Printf("[%s] is valid [%t]", seratoFolder, stat.IsDir())
	}
	log.Printf("Folder search took [%s]", time.Since(start))
	close(pro)
	close(part)
	close(found)
	return seratoFolders
}

func findSeratoFolder(d string) []string {
	files, err := filePathWalkDir(d)
	if err != nil {
		log.Fatal(err)
	}
	return files
}

func filePathWalkDir(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(
		root,
		func(path string, info os.DirEntry, err error) error {
			if (info.IsDir()) && (info.Name() == "_Serato_") {
				files = append(files, path)
			}
			return nil
		},
	)
	return files, err
}
