package server

import (
	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/server/viewrepo"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"path/filepath"
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

func (t *server) GetSubDirs(parentDirPath string, dirs *[]string) (err error) {
	if parentDirPath == "" {
		// Path not specified, return recent used parent paths and root folders
		paths := t.configService.GetState().RecentParentFolders
		paths = append(paths, utils.GetVolumes()...)
		*dirs = paths
		return nil
	}

	// Get sub dirs for the parent dirs
	*dirs, err = utils.GetSubDirs(parentDirPath)
	return
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

	var previousRepo = t.repo()
	if previousRepo != nil {
		// Previous repo was not closed, lets close before starting next
		previousRepo.CloseRepo()
	}

	// Got root working dir path, open repo
	viewRepo := viewrepo.NewViewRepo(t.configService, workingDir)
	t.setRepo(viewRepo)

	viewRepo.StartMonitor()

	// Remember working dir paths to use for "open recent" lists
	parentDir := filepath.Dir(workingDir)
	t.configService.SetState(func(s *config.State) {
		s.RecentFolders = utils.RecentPaths(s.RecentFolders, workingDir, 10)
		s.RecentParentFolders = utils.RecentPaths(s.RecentParentFolders, parentDir, 5)
	})
	return nil
}

func (t *server) CloseRepo(_ api.NoArg, _ api.NoRsp) error {
	t.repo().CloseRepo()
	t.setRepo(nil)
	return nil
}

func (t *server) GetRepoChanges(_ api.NoArg, rsp *[]api.RepoChange) error {
	log.Infof(">")

	repo := t.repo()
	if repo == nil {
		log.Warnf("no repo")
		return nil
	}
	var changes []api.RepoChange
	defer func() { log.Infof("< (%d events)", len(changes)) }()

	// Wait for event or timout
	select {
	case change, ok := <-repo.RepoChangesOut:
		if !ok {
			log.Infof("chan closed")
			return nil
		}
		log.Infof("one event")
		changes = append(changes, change.(api.RepoChange))
	case <-time.After(getChangesTimout):
		// Timeout while whiting for changes, return empty list. Client will retry again
		log.Infof("timeout")
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
		case change, ok := <-repo.RepoChangesOut:
			if !ok {
				log.Infof("chan closed")
				*rsp = changes
				return nil
			}
			changes = append(changes, change.(api.RepoChange))
			log.Infof("more events event (%d events)", len(changes))
		default:
			// no more queued changes,
			log.Infof("no more events (%d events)", len(changes))
			*rsp = changes
			return nil
		}
	}
}

func (t *server) TriggerRefreshRepo(_ api.NoArg, _ api.NoRsp) error {
	t.repo().TriggerRefreshModel()
	return nil
}

func (t *server) TriggerSearch(text string, _ api.NoRsp) error {
	t.repo().TriggerSearch(text)
	return nil
}

func (t *server) GetBranches(args api.GetBranchesArgs, branches *[]api.Branch) error {
	*branches = t.repo().GetBranches(args)
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

func (t *server) Checkout(args api.CheckoutArgs, _ api.NoRsp) error {
	return t.repo().SwitchToBranch(args.Name, args.DisplayName)
}

func (t *server) PushBranch(name string, _ api.NoRsp) error {
	return t.repo().PushBranch(name)
}

func (t *server) PullCurrentBranch(_ api.NoArg, _ api.NoRsp) error {
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

func (t *server) GetCommitDiff(id string, diff *api.CommitDiff) (err error) {
	*diff, err = t.repo().GetCommitDiff(id)
	return
}

func (t *server) Commit(message string, _ api.NoRsp) error {
	return t.repo().Commit(message)
}

func (t *server) setRepo(repo *viewrepo.ViewRepo) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.viewRepo = repo
}

func (t *server) repo() *viewrepo.ViewRepo {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.viewRepo
}
