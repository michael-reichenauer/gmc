package server

import (
	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/server/viewrepo"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const getChangesTimout = 60 * time.Second

type server struct {
	configService *config.Service
	lock          sync.Mutex
	viewRepo      *viewrepo.ViewRepo
}

func NewServer(configService *config.Service) api.Api {
	return &server{configService: configService}
}

func (t *server) GetSubDirs(parentDirPath string) ([]string, error) {
	log.Debugf(">")
	defer log.Debugf("<")
	var paths []string
	if parentDirPath == "" {
		// Path not specified, return recent used parent paths and root folders
		paths = t.configService.GetState().RecentParentFolders
		paths = append(paths, utils.GetVolumes()...)
		return paths, nil
	}

	return t.getSubDirs(parentDirPath), nil
}

func (t *server) OpenRepo(path string) error {
	log.Debugf(">")
	defer log.Debugf("<")
	if path == "" {
		// No path specified, assume current working dir
		path = utils.CurrentDir()
	}
	workingDir, err := git.WorkingDirRoot(path)
	if err != nil {
		// Could not locate a working dir root
		return err
	}

	parentDir := filepath.Dir(workingDir)
	t.configService.SetState(func(s *config.State) {
		s.RecentFolders = utils.RecentItems(s.RecentFolders, workingDir, 10)
		s.RecentParentFolders = utils.RecentItems(s.RecentParentFolders, parentDir, 5)
	})

	viewRepo := viewrepo.NewViewRepo(t.configService, workingDir)
	t.lock.Lock()
	t.viewRepo = viewRepo
	t.lock.Unlock()
	viewRepo.StartMonitor()
	return nil
}

func (t *server) GetRecentWorkingDirs() ([]string, error) {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.configService.GetState().RecentFolders, nil
}

func (t *server) CloseRepo() {
	log.Debugf(">")
	defer log.Debugf("<")
	t.repo().CloseRepo()
	t.lock.Lock()
	t.viewRepo = nil
	t.lock.Unlock()
}

func (t *server) GetChanges() []api.RepoChange {
	repo := t.repo()
	if repo == nil {
		return nil
	}
	var changes []api.RepoChange

	// Wait for event or timout
	select {
	case change, ok := <-repo.RepoChanges:
		if !ok {
			return changes
		}
		changes = append(changes, change)
	case <-time.After(getChangesTimout):
		// Timeout while whiting for changes, return empty list. Client will retry again
		return changes
	}

	repo = t.repo()
	if repo == nil {
		return changes
	}

	// Got some event, check if there are more events and return them as well
	for {
		select {
		case change, ok := <-repo.RepoChanges:
			if !ok {
				return changes
			}
			changes = append(changes, change)
		default:
			// no more queued changes,
			return changes
		}
	}
}

func (t *server) TriggerRefreshModel() {
	log.Debugf(">")
	defer log.Debugf("<")
	t.repo().TriggerRefreshModel()
}

func (t *server) TriggerSearch(text string) {
	log.Debugf(">")
	defer log.Debugf("<")
	t.repo().TriggerSearch(text)
}

func (t *server) GetCommitOpenInBranches(id string) []api.Branch {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.repo().GetCommitOpenInBranches(id)
}

func (t *server) GetCommitOpenOutBranches(id string) []api.Branch {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.repo().GetCommitOpenOutBranches(id)
}

func (t *server) GetCurrentNotShownBranch() (api.Branch, bool) {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.repo().GetCurrentNotShownBranch()
}

func (t *server) GetCurrentBranch() (api.Branch, bool) {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.repo().GetCurrentBranch()
}

func (t *server) GetLatestBranches(shown bool) []api.Branch {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.repo().GetLatestBranches(shown)
}

func (t *server) GetAllBranches(shown bool) []api.Branch {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.repo().GetAllBranches(shown)
}

func (t *server) GetShownBranches(master bool) []api.Branch {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.repo().GetShownBranches(master)
}

func (t *server) ShowBranch(name string) {
	log.Debugf(">")
	defer log.Debugf("<")
	t.repo().ShowBranch(name)
}

func (t *server) HideBranch(name string) {
	log.Debugf(">")
	defer log.Debugf("<")
	t.repo().HideBranch(name)
}

func (t *server) SwitchToBranch(name string, name2 string) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.repo().SwitchToBranch(name, name2)
}

func (t *server) PushBranch(name string) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.repo().PushBranch(name)
}

func (t *server) PullBranch() error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.repo().PullBranch()
}

func (t *server) MergeBranch(name string) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.repo().MergeBranch(name)
}

func (t *server) CreateBranch(name string) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.repo().CreateBranch(name)
}

func (t *server) DeleteBranch(name string) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.repo().DeleteBranch(name)
}

func (t *server) GetCommitDiff(id string) (api.CommitDiff, error) {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.repo().GetCommitDiff(id)
}

func (t *server) Commit(message string) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.repo().Commit(message)
}

func (t *server) getSubDirs(path string) []string {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		// Folder not readable, might be e.g. access denied
		return nil
	}

	var paths []string
	for _, f := range files {
		if !f.IsDir() || f.Name() == "$RECYCLE.BIN" {
			continue
		}
		paths = append(paths, filepath.Join(path, f.Name()))
	}
	// Sort with but ignore case
	sort.SliceStable(paths, func(l, r int) bool {
		return -1 == strings.Compare(strings.ToLower(paths[l]), strings.ToLower(paths[r]))
	})
	return paths
}
func (t *server) repo() *viewrepo.ViewRepo {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.viewRepo
}
