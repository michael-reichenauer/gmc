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
	"github.com/michael-reichenauer/gmc/utils/async"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
)

const getChangesTimeout = 1 * time.Minute

type repoInfo struct {
	repo   *viewrepo.ViewRepoService
	stream observer.Stream
}

type apiServer struct {
	configService *config.Service
	lock          sync.Mutex
	repos         map[string]repoInfo
}

func NewApiServer(configService *config.Service) api.Api {
	return &apiServer{configService: configService, repos: make(map[string]repoInfo)}
}

func (t *apiServer) GetRecentWorkingDirs() ([]string, error) {
	return t.configService.GetState().RecentFolders, nil
}

func (t *apiServer) GetSubDirs(parentDirPath string) ([]string, error) {
	if parentDirPath == "" {
		// Path not specified, return recent used parent paths and root folders
		paths := t.configService.GetState().RecentParentFolders
		paths = append(paths, utils.GetVolumes()...)
		return paths, nil
	}

	// Get sub dirs for the parent dirs
	return utils.GetSubDirs(parentDirPath)
}

func (t *apiServer) OpenRepo(path string) async.Task[string] {
	return async.RunRE(func() (string, error) {
		if path == "" {
			// No path specified, assume current working dir
			path = utils.CurrentDir()
		}
		rootPath, err := git.WorkingTreeRoot(path)
		if err != nil {
			// Could not locate a working dir root
			return "", err
		}

		// Got root working dir path, open repo
		viewRepo := viewrepo.NewViewRepoService(t.configService, rootPath)
		stream := viewRepo.ObserveChanges()
		id := t.storeRepo(viewRepo, stream)

		viewRepo.StartMonitor()

		// Remember working dir paths to use for "open recent" lists
		parentDir := filepath.Dir(rootPath)
		t.configService.SetState(func(s *config.State) {
			s.RecentFolders = utils.RecentPaths(s.RecentFolders, rootPath, 10)
			s.RecentParentFolders = utils.RecentPaths(s.RecentParentFolders, parentDir, 5)
		})

		return id, nil
	})
}

func (t *apiServer) CloseRepo(repoID string) error {
	repo, err := t.repo(repoID)
	if err != nil {
		return err
	}

	t.removeRepo(repoID)
	repo.CloseRepo()
	return nil
}

func (t *apiServer) GetRepoChanges(repoID string) ([]api.RepoChange, error) {
	changesStream, err := t.getStream(repoID)
	if err != nil {
		return []api.RepoChange{}, err
	}

	var changes []api.RepoChange

	// Wait for event or timeout
	select {
	case <-changesStream.Changes():
		changesStream.Next()
		change := changesStream.Value()
		changes = append(changes, change.(api.RepoChange))
	case <-time.After(getChangesTimeout):
		// Timeout while whiting for changes, return empty list. Client will retry again
		return []api.RepoChange{}, nil
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
			return changes, nil
		}
	}
}

func (t *apiServer) TriggerRefreshRepo(repoID string) error {
	repo, err := t.repo(repoID)
	if err != nil {
		return err
	}
	repo.TriggerRefreshModel()
	return nil
}

func (t *apiServer) TriggerSearch(search api.Search) error {
	repo, err := t.repo(search.RepoID)
	if err != nil {
		return err
	}
	repo.TriggerSearch(search.Text)
	return nil
}

func (t *apiServer) GetBranches(args api.GetBranchesReq) ([]api.Branch, error) {
	repo, err := t.repo(args.RepoID)
	if err != nil {
		return []api.Branch{}, err
	}

	return repo.GetBranches(args), nil
}

func (t *apiServer) GetFiles(args api.FilesReq) ([]string, error) {
	repo, err := t.repo(args.RepoID)
	if err != nil {
		return []string{}, err
	}
	files, err := repo.GetFiles(args.Ref)
	if err != nil {
		return []string{}, err
	}
	return files, nil
}

func (t *apiServer) GetAmbiguousBranchBranches(args api.AmbiguousBranchBranchesReq, branches *[]api.Branch) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(args.RepoID)
	if err != nil {
		return err
	}
	*branches = repo.GetAmbiguousBranchBranches(args)
	return nil
}

func (t *apiServer) ShowBranch(name api.BranchName, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(name.RepoID)
	if err != nil {
		return err
	}
	repo.ShowBranch(name.BranchName)
	return nil
}

func (t *apiServer) HideBranch(name api.BranchName, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(name.RepoID)
	if err != nil {
		return err
	}
	repo.HideBranch(name.BranchName)
	return nil
}

func (t *apiServer) Checkout(args api.CheckoutReq, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(args.RepoID)
	if err != nil {
		return err
	}
	return repo.SwitchToBranch(args.Name, args.DisplayName)
}

func (t *apiServer) PushBranch(name api.BranchName, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(name.RepoID)
	if err != nil {
		return err
	}
	return repo.PushBranch(name.BranchName)
}

func (t *apiServer) PullCurrentBranch(repoID string, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(repoID)
	if err != nil {
		return err
	}
	return repo.PullCurrentBranch()
}

func (t *apiServer) PullBranch(name api.BranchName, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(name.RepoID)
	if err != nil {
		return err
	}
	return repo.PullBranch(name.BranchName)
}

func (t *apiServer) MergeBranch(name api.BranchName, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(name.RepoID)
	if err != nil {
		return err
	}
	return repo.MergeBranch(name.BranchName)
}

func (t *apiServer) CreateBranch(name api.BranchName, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(name.RepoID)
	if err != nil {
		return err
	}
	return repo.CreateBranch(name.BranchName)
}

func (t *apiServer) DeleteBranch(name api.BranchName, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(name.RepoID)
	if err != nil {
		return err
	}
	return repo.DeleteBranch(name.BranchName)
}

func (t *apiServer) GetCommitDiff(info api.CommitDiffInfoReq) (api.CommitDiff, error) {
	repo, err := t.repo(info.RepoID)
	if err != nil {
		return api.CommitDiff{}, err
	}

	return repo.GetCommitDiff(info.CommitID)
}

func (t *apiServer) GetFileDiff(info api.FileDiffInfoReq) ([]api.CommitDiff, error) {
	repo, err := t.repo(info.RepoID)
	if err != nil {
		return []api.CommitDiff{}, err
	}

	return repo.GetFileDiff(info.Path)
}

func (t *apiServer) GetCommitDetails(args api.CommitDetailsReq) (api.CommitDetailsRsp, error) {
	repo, err := t.repo(args.RepoID)
	if err != nil {
		return api.CommitDetailsRsp{}, err
	}

	return repo.GetCommitDetails(args.CommitID)
}

func (t *apiServer) Commit(info api.CommitInfoReq, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(info.RepoID)
	if err != nil {
		return err
	}
	return repo.Commit(info.Message)
}

func (t *apiServer) UndoCommit(info api.IdReq, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(info.RepoID)
	if err != nil {
		return err
	}
	return repo.UndoCommit(info.Id)
}

func (t *apiServer) UndoUncommittedFileChanges(info api.FilesReq, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(info.RepoID)
	if err != nil {
		return err
	}
	return repo.UndoUncommittedFileChanges(info.Ref)
}

func (t *apiServer) UncommitLastCommit(repoID string, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(repoID)
	if err != nil {
		return err
	}
	return repo.UncommitLastCommit()
}

func (t *apiServer) UndoAllUncommittedChanges(repoID string, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(repoID)
	if err != nil {
		return err
	}
	return repo.UndoAllUncommittedChanges()
}

func (t *apiServer) CleanWorkingFolder(repoID string, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(repoID)
	if err != nil {
		return err
	}
	return repo.CleanWorkingFolder()
}

func (t *apiServer) SetAsParentBranch(name api.SetParentReq, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	repo, err := t.repo(name.RepoID)
	if err != nil {
		return err
	}

	return repo.SetAsParentBranch(name.BranchName, name.ParentName)
}

func (t *apiServer) UnsetAsParentBranch(name api.BranchName, _ api.NoRsp) error {
	log.Infof(">")
	defer log.Infof("<")
	log.Infof("Set as parent %q", name)
	repo, err := t.repo(name.RepoID)
	if err != nil {
		return err
	}

	return repo.UnsetAsParentBranch(name.BranchName)
}

func (t *apiServer) storeRepo(repo *viewrepo.ViewRepoService, stream observer.Stream) string {
	t.lock.Lock()
	defer t.lock.Unlock()
	id := uuid.New().String()
	t.repos[id] = repoInfo{repo: repo, stream: stream}
	return id
}

func (t *apiServer) removeRepo(id string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	delete(t.repos, id)
}

func (t *apiServer) repo(id string) (*viewrepo.ViewRepoService, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	r, ok := t.repos[id]
	if !ok {
		return nil, fmt.Errorf("repo not open")
	}
	return r.repo, nil
}

func (t *apiServer) getStream(id string) (observer.Stream, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	r, ok := t.repos[id]
	if !ok {
		return nil, fmt.Errorf("repo not open")
	}
	return r.stream, nil
}
