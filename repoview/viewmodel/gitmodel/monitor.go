package gitmodel

import (
	"github.com/fsnotify/fsnotify"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
	"os"
	"path/filepath"
	"strings"
)

// ssk.
type monitor struct {
	StatusChange  chan interface{}
	RepoChange    chan interface{}
	watcher       *fsnotify.Watcher
	repoPath      string
	gitPath       string
	refsPath      string
	headPath      string
	fetchHeadPath string
	quit          chan chan<- interface{}
}

func newMonitor(repoPath string) *monitor {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	return &monitor{
		watcher:       watcher,
		repoPath:      repoPath,
		gitPath:       filepath.Join(repoPath, ".git"),
		refsPath:      filepath.Join(repoPath, ".git", "refs"),
		headPath:      filepath.Join(repoPath, ".git", "HEAD"),
		fetchHeadPath: filepath.Join(repoPath, ".git", "FETCH_HEAD"),
		StatusChange:  make(chan interface{}),
		RepoChange:    make(chan interface{}),
		quit:          make(chan chan<- interface{}),
	}
}

func (h *monitor) Start() error {
	go h.monitorFolderRoutine()
	go filepath.Walk(h.repoPath, h.watchDir)
	return nil
}

func (h *monitor) Close() {
	h.watcher.Close()
	closed := make(chan interface{})
	h.quit <- closed
	<-closed
}

func (h *monitor) monitorFolderRoutine() {
	var err error
	for {
		select {
		case event := <-h.watcher.Events:
			if h.isIgnore(event.Name) {
				//fmt.Printf("ignoring: %s\n", event.Name)
			} else if h.isRepoChange(event.Name) {
				//log.Infof("Repo change: %s", event.Name)
				select {
				case h.RepoChange <- nil:
				default:
				}
			} else if h.isStatusChange(event.Name) {
				//log.Infof("Status change: %s", event.Name)
				select {
				case h.StatusChange <- nil:
				default:
				}
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

func (h *monitor) watchDir(path string, fi os.FileInfo, err error) error {
	if fi.Mode().IsDir() {
		return h.watcher.Add(path)
	}
	return nil
}
func (h *monitor) isIgnore(path string) bool {
	if utils.DirExists(path) {
		return true
	}
	return false
}
func (h *monitor) isStatusChange(path string) bool {
	if strings.HasPrefix(path, h.gitPath) {
		return false
	}
	return true
}

func (h *monitor) isRepoChange(path string) bool {
	if strings.HasSuffix(path, ".lock") {
		return false
	}
	if path == h.fetchHeadPath {
		return false
	}
	if path == h.headPath {
		return true
	}
	if strings.HasPrefix(path, h.refsPath) {
		return true
	}
	return false
}
