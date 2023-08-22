package dirmonitor

import (
	"log"
	"os"
	fp "path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// filepath represents the full filepath (includes the base dir)
type fileListenFunc func(filepath string) error

// filename represents just the filename (excludes the base dir)
type dirListenFunc func(filename string) error

// DirMonitor monitor the 'write' events of all the files in a specified directory (not recursively)
// different files can bind different listen functions respectively
type DirMonitor struct {
	dirPath       string
	dirChildPaths []string
	watcher       *fsnotify.Watcher
	flfMapping    map[string]fileListenFunc
	dlf           dirListenFunc
}

// NewDirMonitor initiate with a specified directory path
func NewDirMonitor(dirPath string) (*DirMonitor, error) {
	dm := DirMonitor{
		dirPath:    fp.Clean(dirPath),
		flfMapping: make(map[string]fileListenFunc),
	}
	return &dm, nil
}

// NewDirMonitorRecursively initiate with a specified directory path and also including dirPath all child path
func NewDirMonitorRecursively(dirPath string) (*DirMonitor, error) {
	cleanDirPath := fp.Clean(dirPath)
	dm := DirMonitor{
		dirPath:       cleanDirPath,
		dirChildPaths: listChildPath(cleanDirPath),
		flfMapping:    make(map[string]fileListenFunc),
	}
	return &dm, nil
}

func listChildPath(root string) []string {
	childPathList := make([]string, 0)
	_ = fp.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			childPathList = append(childPathList, path)
		}
		return nil
	})
	return childPathList
}

// Bind bind file to listen function
// must be called before StartWatch
func (dm *DirMonitor) Bind(filename string, flf fileListenFunc) {
	dm.flfMapping[filename] = flf
}

// BindAndExec bind file to listen function, and then execute the function
// must be called before StartWatch
func (dm *DirMonitor) BindAndExec(filename string, flf fileListenFunc) error {
	dm.flfMapping[filename] = flf
	return flf(fp.Join(dm.dirPath, filename))
}

// BindAny bind all files to listen function
// must be called before StartWatch
func (dm *DirMonitor) BindAny(dlf dirListenFunc) {
	dm.dlf = dlf
}

// BindAnyCsv bind all csv files to listen function
// must be called before StartWatch
func (dm *DirMonitor) BindAnyCsv(dlf dirListenFunc) {
	csvDlf := func(filename string) error {
		pathSlices := strings.Split(filename, ".")
		if len(pathSlices) != 2 || pathSlices[1] != "csv" {
			return nil
		}
		return dlf(filename)
	}
	dm.dlf = csvDlf
}

// StartWatch start the monitor
// must be called once after all the Bind & BindAndExec methods have been called
func (dm *DirMonitor) StartWatch() error {
	var err error
	if dm.watcher, err = fsnotify.NewWatcher(); err != nil {
		return err
	}

	go dm.watchEvents()

	err = dm.watcher.Add(dm.dirPath)
	if err != nil {
		return err
	}

	for _, childDir := range dm.dirChildPaths {
		err = dm.watcher.Add(childDir)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dm *DirMonitor) watchEvents() {
	for {
		select {
		case event, ok := <-dm.watcher.Events:
			if !ok {
				log.Println("watcher closed")
				return
			}
			log.Println("event:", event)
			if event.Op&fsnotify.Write == fsnotify.Write {
				if err := dm.dispatch(event.Name); err != nil {
					log.Println(err)
				}
			}
		case err, ok := <-dm.watcher.Errors:
			if !ok {
				log.Println("watcher closed")
				return
			}
			log.Println("error:", err)
		}
	}
}

func (dm *DirMonitor) dispatch(eventName string) error {
	filename, err := fp.Rel(dm.dirPath, eventName)
	if err != nil {
		return err
	}
	if flf, ok := dm.flfMapping[filename]; ok {
		if err := flf(eventName); err != nil {
			return err
		}
	}
	if dm.dlf != nil {
		return dm.dlf(filename)
	}
	return nil
}
