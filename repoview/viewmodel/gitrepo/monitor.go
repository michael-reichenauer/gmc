package gitrepo

import (
	"github.com/fsnotify/fsnotify"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
	"os"
	"path/filepath"
	"strings"
)

type Ignorer interface {
	IsIgnored(path string) bool
}

type monitor struct {
	RepoChange   chan struct{}
	StatusChange chan struct{}

	watcher        *fsnotify.Watcher
	rootFolderPath string
	ignorer        Ignorer
	quit           chan chan<- struct{}
}

func newMonitor(rootFolderPath string, ignorer Ignorer) *monitor {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(log.Fatal(err))
	}
	return &monitor{
		RepoChange:     make(chan struct{}),
		StatusChange:   make(chan struct{}),
		watcher:        watcher,
		rootFolderPath: rootFolderPath,
		ignorer:        ignorer,
		quit:           make(chan chan<- struct{}),
	}
}

func (h *monitor) Start() error {
	go h.monitorFolderRoutine()
	go h.addWatchFoldersRecursively(h.rootFolderPath)
	return nil
}

func (h *monitor) addWatchFoldersRecursively(path string) {
	filepath.Walk(path, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.Mode().IsDir() {
			return h.watcher.Add(path)
		}
		return nil
	})
}

func (h *monitor) Close() {
	h.watcher.Close()
	closed := make(chan struct{})
	h.quit <- closed
	<-closed
}

func (h *monitor) monitorFolderRoutine() {
	gitPath := filepath.Join(h.rootFolderPath, ".git")
	refsPath := filepath.Join(gitPath, "refs")
	headPath := filepath.Join(gitPath, "HEAD")
	fetchHeadPath := filepath.Join(gitPath, "FETCH_HEAD")

	var err error
	for {
		select {
		case event := <-h.watcher.Events:
			if h.isNewFolder(event) {
				log.Infof("New folder detected: %q", event.Name)
				go h.addWatchFoldersRecursively(event.Name)
			}
			if h.isIgnored(event.Name) {
				//log.Infof("ignoring: %s", event.Name)
			} else if h.isRepoChange(event.Name, fetchHeadPath, headPath, refsPath) {
				//log.Infof("Repo change: %s", event.Name)
				h.postRepoChange()
			} else if h.isStatusChange(event.Name, gitPath) {
				//log.Infof("Status change: %s", event.Name)
				h.postStatusChange()
			} else {
				// fmt.Printf("ignoring: %s\n", event.Name)
			}
		case err = <-h.watcher.Errors:
			log.Warnf("ERROR %v", err)
		case closed := <-h.quit:
			close(h.RepoChange)
			close(h.StatusChange)
			close(closed)
			return
		}
	}
}
func (h *monitor) postRepoChange() {
	select {
	case h.RepoChange <- struct{}{}:
	default:
		// Ignore if no listener
	}
}

func (h *monitor) postStatusChange() {
	select {
	case h.StatusChange <- struct{}{}:
	default:
		// Ignore if no listener
	}
}

func (h *monitor) isIgnored(path string) bool {
	if utils.DirExists(path) {
		return true
	}
	if h.ignorer != nil && h.ignorer.IsIgnored(path) {
		return true
	}
	return false
}

func (h *monitor) isStatusChange(path, gitPath string) bool {
	if strings.HasPrefix(path, gitPath) {
		return false
	}
	return true
}

func (h *monitor) isRepoChange(path, fetchHeadPath, headPath, refsPath string) bool {
	if strings.HasSuffix(path, ".lock") {
		return false
	}
	if path == fetchHeadPath {
		return false
	}
	if path == headPath {
		return true
	}
	if strings.HasPrefix(path, refsPath) {
		return true
	}
	return false
}

func (h *monitor) isNewFolder(event fsnotify.Event) bool {
	if event.Op != fsnotify.Create {
		return false
	}
	if utils.DirExists(event.Name) {
		return true
	}
	return false
}
