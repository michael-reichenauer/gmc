package gitrepo

import (
	"context"
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

type changeType int

const (
	noChange changeType = iota
	repoChange
	statusChange
)

type monitor struct {
	Changes chan changeType

	watcher        *fsnotify.Watcher
	rootFolderPath string
	ignorer        Ignorer
}

func newMonitor(rootFolderPath string, ignorer Ignorer) *monitor {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(log.Fatal(err))
	}
	return &monitor{
		Changes:        make(chan changeType),
		watcher:        watcher,
		rootFolderPath: rootFolderPath,
		ignorer:        ignorer,
	}
}

func (h *monitor) Start(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		h.watcher.Close()
	}()
	go h.monitorFolderRoutine(ctx)
	go h.addWatchFoldersRecursively(ctx, h.rootFolderPath)
	return nil
}

func (h *monitor) addWatchFoldersRecursively(ctx context.Context, path string) {
	filepath.Walk(path, func(path string, fi os.FileInfo, err error) error {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		if err != nil {
			return err
		}
		if fi.Mode().IsDir() {
			return h.watcher.Add(path)
		}
		return nil
	})
}

func (h *monitor) monitorFolderRoutine(ctx context.Context) {
	gitPath := filepath.Join(h.rootFolderPath, ".git")
	refsPath := filepath.Join(gitPath, "refs")
	headPath := filepath.Join(gitPath, "HEAD")
	fetchHeadPath := filepath.Join(gitPath, "FETCH_HEAD")
	defer close(h.Changes)

	for event := range h.watcher.Events {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if h.isNewFolder(event) {
			log.Infof("New folder detected: %q", event.Name)
			go h.addWatchFoldersRecursively(ctx, event.Name)
			continue
		}
		if h.isIgnored(event.Name) {
			// log.Infof("ignoring: %s", event.Name)
		} else if h.isRepoChange(event.Name, fetchHeadPath, headPath, refsPath) {
			// log.Infof("Repo change: %s", event.Name)
			h.Changes <- repoChange
		} else if h.isStatusChange(event.Name, gitPath) {
			// log.Infof("Status change: %s", event.Name)
			h.Changes <- statusChange
		} else {
			// fmt.Printf("ignoring: %s\n", event.Name)
		}
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
