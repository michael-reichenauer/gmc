package server

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/imkira/go-observer"
	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/server/viewrepo"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
)

const getChangesTimout = 1 * time.Minute

type repoInfo struct {
	repo   *viewrepo.ViewRepo
	stream observer.Stream
}

type server struct {
	configService *config.Service
	lock          sync.Mutex
	repos         map[string]repoInfo
}

func NewServer(configService *config.Service) api.Api {
	return &server{configService: configService, repos: make(map[string]repoInfo)}
}

func (t *server) GetRecentWorkingDirs(_ api.NoArg, rsp *[]string) error {
	*rsp = t.configService.GetState().RecentFolders
	return nil
}

func (t *server) GetSubDirs(parentDirPath string, dirs *[]string) (err error) {
	log.Infof(">")
	defer log.Infof("<")
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

func (t *server) OpenRepo(path string, repoID *string) error {
	log.Infof(">")
	defer log.Infof("<")
	if path == "" {
		// No path specified, assume current working dir
		path = utils.CurrentDir()
	}
	workingDir, err := git.WorkingDirRoot(path)
	if err != nil {
		// Could not locate a working dir root
		return err
	}

	// Got root working dir path, open repo
	viewRepo := viewrepo.NewViewRepo(t.configService, workingDir)
	stream := viewRepo.ObserveChanges()
	id := t.storeRepo(viewRepo, stream)

	viewRepo.StartMonitor()

	// Remember working dir paths to use for "open recent" lists
	parentDir := filepath.Dir(workingDir)
	t.configService.SetState(func(s *config.State) {
		s.RecentFolders = utils.RecentPaths(s.RecentFolders, workingDir, 10)
		s.RecentParentFolders = utils.RecentPaths(s.RecentParentFolders, parentDir, 5)
	})
	*repoID = id
	return nil
}

func (t *server) CloseRepo(repoID string, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(repoID)
	if err != nil {
		return err
	}
	t.removeRepo(repoID)
	repo.CloseRepo()
	return nil
}

func (t *server) GetRepoChanges(repoID string, rsp *[]api.RepoChange) error {
	log.Infof(">")
	defer log.Infof("<")
	changesStream, err := t.getStream(repoID)
	if err != nil {
		return err
	}

	var changes []api.RepoChange
	defer func() { log.Infof("GetRepoChanges: < (%d events)", len(changes)) }()

	// Wait for event or timeout
	select {
	case <-changesStream.Changes():
		changesStream.Next()
		change := changesStream.Value()
		log.Debugf("one event")
		changes = append(changes, change.(api.RepoChange))
	case <-time.After(getChangesTimout):
		// Timeout while whiting for changes, return empty list. Client will retry again
		log.Debugf("timeout")
		return nil
	}

	// Got some event, check if there are more events and return them as well
	for {
		select {
		case <-changesStream.Changes():
			changesStream.Next()
			change := changesStream.Value()
			changes = append(changes, change.(api.RepoChange))
			log.Debugf("more events event (%d events)", len(changes))
		default:
			// no more queued changes,
			log.Debugf("no more events (%d events)", len(changes))
			*rsp = changes
			return nil
		}
	}
}

func (t *server) TriggerRefreshRepo(repoID string, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(repoID)
	if err != nil {
		return err
	}
	repo.TriggerRefreshModel()
	return nil
}

func (t *server) TriggerSearch(search api.Search, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(search.RepoID)
	if err != nil {
		return err
	}
	repo.TriggerSearch(search.Text)
	return nil
}

func (t *server) GetBranches(args api.GetBranches, branches *[]api.Branch) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(args.RepoID)
	if err != nil {
		return err
	}
	*branches = repo.GetBranches(args)
	return nil
}

func (t *server) ShowBranch(name api.BranchName, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(name.RepoID)
	if err != nil {
		return err
	}
	repo.ShowBranch(name.BranchName)
	return nil
}

func (t *server) HideBranch(name api.BranchName, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(name.RepoID)
	if err != nil {
		return err
	}
	repo.HideBranch(name.BranchName)
	return nil
}

func (t *server) Checkout(args api.Checkout, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(args.RepoID)
	if err != nil {
		return err
	}
	return repo.SwitchToBranch(args.Name, args.DisplayName)
}

func (t *server) PushBranch(name api.BranchName, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(name.RepoID)
	if err != nil {
		return err
	}
	return repo.PushBranch(name.BranchName)
}

func (t *server) PullCurrentBranch(repoID string, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(repoID)
	if err != nil {
		return err
	}
	return repo.PullBranch()
}

func (t *server) MergeBranch(name api.BranchName, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(name.RepoID)
	if err != nil {
		return err
	}
	return repo.MergeBranch(name.BranchName)
}

func (t *server) CreateBranch(name api.BranchName, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(name.RepoID)
	if err != nil {
		return err
	}
	return repo.CreateBranch(name.BranchName)
}

func (t *server) DeleteBranch(name api.BranchName, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(name.RepoID)
	if err != nil {
		return err
	}
	return repo.DeleteBranch(name.BranchName)
}

func (t *server) GetCommitDiff(info api.CommitDiffInfo, diff *api.CommitDiff) (err error) {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(info.RepoID)
	if err != nil {
		return err
	}
	*diff, err = repo.GetCommitDiff(info.CommitID)
	return
}

func (t *server) Commit(info api.CommitInfo, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(info.RepoID)
	if err != nil {
		return err
	}
	return repo.Commit(info.Message)
}

func (t *server) storeRepo(repo *viewrepo.ViewRepo, stream observer.Stream) string {
	t.lock.Lock()
	defer t.lock.Unlock()
	id := uuid.New().String()
	t.repos[id] = repoInfo{repo: repo, stream: stream}
	return id
}

func (t *server) removeRepo(id string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	delete(t.repos, id)
}

func (t *server) repo(id string) (*viewrepo.ViewRepo, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	r, ok := t.repos[id]
	if !ok {
		return nil, fmt.Errorf("repo not open")
	}
	return r.repo, nil
}

func (t *server) getStream(id string) (observer.Stream, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	r, ok := t.repos[id]
	if !ok {
		return nil, fmt.Errorf("repo not open")
	}
	return r.stream, nil
}
