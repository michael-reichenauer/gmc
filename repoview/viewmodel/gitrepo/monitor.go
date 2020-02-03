package gitrepo

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
	StatusChange chan interface{}
	RepoChange   chan interface{}
	watcher      *fsnotify.Watcher
	// gitPath       string
	// refsPath      string
	// headPath      string
	// fetchHeadPath string
	quit chan chan<- interface{}
	//	isIgnoreFunc  func(path string) bool
}

func newMonitor() *monitor {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(log.Fatal(err))
	}
	return &monitor{
		watcher: watcher,
		// gitPath:       filepath.Join(repoPath, ".git"),
		// refsPath:      filepath.Join(repoPath, ".git", "refs"),
		// headPath:      filepath.Join(repoPath, ".git", "HEAD"),
		// fetchHeadPath: filepath.Join(repoPath, ".git", "FETCH_HEAD"),
		StatusChange: make(chan interface{}),
		RepoChange:   make(chan interface{}),
		quit:         make(chan chan<- interface{}),
		//isIgnoreFunc:  isIgnore,
	}
}

func (h *monitor) Start(repoPath string, isIgnore func(path string) bool) error {

	go h.monitorFolderRoutine(repoPath, isIgnore)

	go filepath.Walk(repoPath, h.watchDir)
	return nil
}

func (h *monitor) Close() {
	h.watcher.Close()
	closed := make(chan interface{})
	h.quit <- closed
	<-closed
}

func (h *monitor) monitorFolderRoutine(repoPath string, isIgnore func(path string) bool) {
	gitPath := filepath.Join(repoPath, ".git")
	refsPath := filepath.Join(repoPath, ".git", "refs")
	headPath := filepath.Join(repoPath, ".git", "HEAD")
	fetchHeadPath := filepath.Join(repoPath, ".git", "FETCH_HEAD")

	var err error
	for {
		select {
		case event := <-h.watcher.Events:
			if h.isIgnore(event.Name, isIgnore) {
				//log.Infof("ignoring: %s", event.Name)
			} else if h.isRepoChange(event.Name, fetchHeadPath, headPath, refsPath) {
				//log.Infof("Repo change: %s", event.Name)
				select {
				case h.RepoChange <- nil:
				default:
				}
			} else if h.isStatusChange(event.Name, gitPath) {
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
	if err != nil {
		return err
	}
	if fi.Mode().IsDir() {
		return h.watcher.Add(path)
	}
	return nil
}

func (h *monitor) isIgnore(path string, isIgnore func(path string) bool) bool {
	if utils.DirExists(path) {
		return true
	}
	if isIgnore != nil && isIgnore(path) {
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
