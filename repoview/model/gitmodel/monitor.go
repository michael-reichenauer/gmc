package gitmodel

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
	"os"
	"path/filepath"
	"strings"
)

// ssk.
type monitor struct {
	FileChange chan interface{}
	RepoChange chan interface{}
	watcher    *fsnotify.Watcher
	repoPath   string
	gitPath    string
	refsPath   string
	fetchHead  string
	quit       chan chan<- interface{}
}

func newMonitor(repoPath string) *monitor {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	return &monitor{
		watcher:    watcher,
		repoPath:   repoPath,
		gitPath:    filepath.Join(repoPath, ".git"),
		refsPath:   filepath.Join(repoPath, ".git", "refs"),
		fetchHead:  filepath.Join(repoPath, ".git", "FETCH_HEAD"),
		FileChange: make(chan interface{}),
		RepoChange: make(chan interface{}),
		quit:       make(chan chan<- interface{}),
	}
}

func (h *monitor) Start() error {
	go h.monitorRoutine()
	return filepath.Walk(h.repoPath, h.watchDir)
}

func (h *monitor) Close() {
	h.watcher.Close()
	closed := make(chan interface{})
	h.quit <- closed
	<-closed
}

func (h *monitor) monitorRoutine() {
	var err error
	for {
		select {
		case event := <-h.watcher.Events:
			if h.isIgnore(event.Name) {
				fmt.Printf("ignoring: %s\n", event.Name)
			} else if h.isRepoChange(event.Name) {
				fmt.Printf("Repo change: %s\n", event.Name)
				select {
				case h.RepoChange <- nil:
				default:
				}
			} else if h.isStatusChange(event.Name) {
				fmt.Printf("Status change: %s\n", event.Name)
				select {
				case h.FileChange <- nil:
				default:
				}
			} else {
				fmt.Printf("ignoring: %s\n", event.Name)
			}
		case err = <-h.watcher.Errors:
			fmt.Println("ERROR", err)
		case closed := <-h.quit:
			close(h.FileChange)
			close(h.FileChange)
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
	if path == h.fetchHead {
		return true
	}
	if strings.HasPrefix(path, h.refsPath) {
		return true
	}
	return false
}
