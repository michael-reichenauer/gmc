package server

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/server/viewrepo"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/git"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const getChangesTimout = 5 * time.Minute

type server struct {
	configService *config.Service
	lock          sync.Mutex
	viewRepo      *viewrepo.ViewRepo
}

func NewServer(configService *config.Service) api.Api {
	return &server{configService: configService}
}

func (t *server) GetRecentWorkingDirs(_ api.NoArg, rsp *[]string) error {
	*rsp = t.configService.GetState().RecentFolders
	return nil
}

func (t *server) GetSubDirs(parentDirPath string, dirs *[]string) error {
	var paths []string
	if parentDirPath == "" {
		// Path not specified, return recent used parent paths and root folders
		paths = t.configService.GetState().RecentParentFolders
		paths = append(paths, utils.GetVolumes()...)
		*dirs = paths
		return nil
	}
	*dirs = t.getSubDirs(parentDirPath)
	return nil
}

func (t *server) OpenRepo(path string, _ api.NoRsp) error {
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

func (t *server) CloseRepo(_ api.NoArg, _ api.NoRsp) error {
	t.repo().CloseRepo()
	t.lock.Lock()
	t.viewRepo = nil
	t.lock.Unlock()
	return nil
}

func (t *server) GetChanges(_ api.NoArg, rsp *[]api.RepoChange) error {
	repo := t.repo()
	if repo == nil {
		return nil
	}
	var changes []api.RepoChange

	// Wait for event or timout
	select {
	case change, ok := <-repo.RepoChanges:
		if !ok {
			return nil
		}
		changes = append(changes, change)
	case <-time.After(getChangesTimout):
		// Timeout while whiting for changes, return empty list. Client will retry again
		return nil
	}

	repo = t.repo()
	if repo == nil {
		*rsp = changes
		return nil
	}

	// Got some event, check if there are more events and return them as well
	for {
		select {
		case change, ok := <-repo.RepoChanges:
			if !ok {
				*rsp = changes
				return nil

			}
			changes = append(changes, change)
		default:
			// no more queued changes,
			*rsp = changes
			return nil
		}
	}
}

func (t *server) TriggerRefreshModel(_ api.NoArg, _ api.NoRsp) error {
	t.repo().TriggerRefreshModel()
	return nil
}

func (t *server) TriggerSearch(text string, _ api.NoRsp) error {
	t.repo().TriggerSearch(text)
	return nil
}

func (t *server) GetCommitBranches(id string, rsp *[]api.Branch) error {
	b := t.repo().GetCommitBranches(id)
	if b != nil {
		*rsp = b
	}
	return nil
}

func (t *server) GetCurrentNotShownBranch(_ api.NoArg, rsp *api.Branch) error {
	b, ok := t.repo().GetCurrentNotShownBranch()
	if !ok {
		return fmt.Errorf("branch not found")
	}
	*rsp = b
	return nil
}

func (t *server) GetCurrentBranch(_ api.NoArg, rsp *api.Branch) error {
	b, ok := t.repo().GetCurrentBranch()
	if !ok {
		return fmt.Errorf("branch not found")
	}
	*rsp = b
	return nil
}

func (t *server) GetLatestBranches(shown bool, rsp *[]api.Branch) error {
	b := t.repo().GetLatestBranches(shown)
	if b != nil {
		*rsp = b
	}
	return nil
}

func (t *server) GetAllBranches(shown bool, rsp *[]api.Branch) error {
	b := t.repo().GetAllBranches(shown)
	if b != nil {
		*rsp = b
	}
	return nil
}

func (t *server) GetShownBranches(master bool, rsp *[]api.Branch) error {
	b := t.repo().GetShownBranches(master)
	if b != nil {
		*rsp = b
	}
	return nil
}

func (t *server) ShowBranch(name string, _ api.NoRsp) error {
	t.repo().ShowBranch(name)
	return nil
}

func (t *server) HideBranch(name string, _ api.NoRsp) error {
	t.repo().HideBranch(name)
	return nil
}

func (t *server) SwitchToBranch(args api.SwitchArgs, _ api.NoRsp) error {
	return t.repo().SwitchToBranch(args.Name, args.DisplayName)
}

func (t *server) PushBranch(name string, _ api.NoRsp) error {
	return t.repo().PushBranch(name)
}

func (t *server) PullBranch(_ api.NoArg, _ api.NoRsp) error {
	return t.repo().PullBranch()
}

func (t *server) MergeBranch(name string, _ api.NoRsp) error {
	return t.repo().MergeBranch(name)
}

func (t *server) CreateBranch(name string, _ api.NoRsp) error {
	return t.repo().CreateBranch(name)
}

func (t *server) DeleteBranch(name string, _ api.NoRsp) error {
	return t.repo().DeleteBranch(name)
}

func (t *server) GetCommitDiff(id string, diff *api.CommitDiff) error {
	d, err := t.repo().GetCommitDiff(id)
	*diff = d
	return err
}

func (t *server) Commit(message string, _ api.NoRsp) error {
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
