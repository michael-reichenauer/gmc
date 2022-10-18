package augmented

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/timer"
	"github.com/samber/lo"
)

type RepoService interface {
	RepoChanges() chan RepoChange
	RepoPath() string
	TriggerManualRefresh()

	StartMonitor(ctx context.Context)

	GetCommitDiff(id string) (git.CommitDiff, error)
	GetFileDiff(path string) ([]git.CommitDiff, error)
	GetFiles(ref string) ([]string, error)

	SwitchToBranch(name string) error
	Commit(commit string) error
	PushBranch(name string) error
	CreateBranch(name string) error
	MergeBranch(name string) error
	DeleteRemoteBranch(name string) error
	DeleteLocalBranch(name string) error
	PullCurrentBranch() error
	PullBranch(name string) error

	GetFreshRepo() (Repo, error)
	SetAsParentBranch(b *Branch, name string) error
	UnsetAsParentBranch(name string) error
}

type RepoChange struct {
	IsStarting bool
	Repo       Repo
	Error      error
}

var metaDataKey = "data"

type gitRepo struct {
	RootPath string
	Commits  []git.Commit
	Branches []git.Branch
	Status   git.Status
	Tags     []git.Tag
	MetaData MetaData
}

type repoService struct {
	branchesService *branchesService
	folderMonitor   *monitor

	repoChanges   chan RepoChange
	git           git.Git
	repo          chan Repo
	manualRefresh chan struct{}
}

const (
	fetchInterval = 10 * time.Minute
	batchInterval = 1 * time.Second
	partialMax    = 30000 // Max number of commits to handle
)

func NewRepoService(rootPath string) RepoService {
	g := git.New(rootPath)

	return &repoService{
		branchesService: newBranchesService(),
		git:             g,
		folderMonitor:   newMonitor(rootPath, g),
		repoChanges:     make(chan RepoChange, 1),
		repo:            make(chan Repo, 1),
		manualRefresh:   make(chan struct{}, 1),
	}
}

func (s *repoService) RepoChanges() chan RepoChange {
	return s.repoChanges
}

func (s *repoService) RepoPath() string {
	return s.git.RepoPath()
}

func (s *repoService) StartMonitor(ctx context.Context) {
	go s.monitorRoutine(ctx)
	go s.fetchRoutine(ctx)
}

func (s *repoService) GetCommitDiff(id string) (git.CommitDiff, error) {
	return s.git.CommitDiff(id)
}

func (s *repoService) GetFileDiff(path string) ([]git.CommitDiff, error) {
	return s.git.FileDiff(path)
}

func (s *repoService) SwitchToBranch(name string) error {
	return s.git.Checkout(name)
}

func (s *repoService) Commit(message string) error {
	return s.git.Commit(message)
}

func (s *repoService) PushBranch(name string) error {
	return s.git.PushBranch(name)
}

func (s *repoService) PullCurrentBranch() error {
	return s.git.PullCurrentBranch()
}

func (s *repoService) PullBranch(name string) error {
	return s.git.PullBranch(name)
}

func (s *repoService) MergeBranch(name string) error {
	return s.git.MergeBranch(name)
}

func (s *repoService) CreateBranch(name string) error {
	return s.git.CreateBranch(name)
}

func (s *repoService) GetFiles(ref string) ([]string, error) {
	return s.git.GetFiles(ref)
}

func (s *repoService) DeleteRemoteBranch(name string) error {
	return s.git.DeleteRemoteBranch(name)
}

func (s *repoService) DeleteLocalBranch(name string) error {
	return s.git.DeleteLocalBranch(name)
}

func (s *repoService) TriggerManualRefresh() {
	select {
	case s.manualRefresh <- struct{}{}:
	default:
		log.Infof("TriggerManualRefresh full")
	}
}

func (s *repoService) monitorRoutine(ctx context.Context) {
	log.Infof("monitorRoutine start.")
	defer close(s.repoChanges)
	defer log.Infof("Closed monitor of %s", s.git.RepoPath())
	s.folderMonitor.Start(ctx)

	hasRepo := false
	var repo Repo
	var wait = time.After(batchInterval)
	change := noChange

	for {
		select {
		case <-wait:
			// Some time has passed, check if there is a repo or status change event to act on
			// This is to avoid that several change events in a short interval is batched
			if change != noChange {
				log.Debugf("waited for %v", change)
				if change == statusChange && hasRepo {
					s.triggerStatus(ctx, repo)
				} else {
					s.triggerRepo(ctx, false)
				}
				change = noChange
			}
			wait = time.After(batchInterval)

		case changeEvent, ok := <-s.folderMonitor.Changes:
			// Received changed repo or status event
			if !ok {
				// No more change events, closing this repo
				return
			}
			if changeEvent == statusChange && !hasRepo {
				// Status change, but we have not yet a repo, ignore status until repo exist
				log.Infof("No repo for status change %v", changeEvent)
				break
			}
			if change == noChange {
				// First change event, ensure we do wait a while before acting on the event
				// This is to avoid that several change events in a short interval is batched
				wait = time.After(batchInterval)
				log.Infof("Got repo start change ...")
				select {
				case s.repoChanges <- RepoChange{IsStarting: true}:
				case <-ctx.Done():
					return
				}
			}
			if change != repoChange {
				// Do not downgrade from repo to status event, status is included in repo change
				change = changeEvent
			}

		case repo = <-s.repo:
			// Received new repo, notify listeners
			log.Debugf("Received repo")
			hasRepo = true
			select {
			case s.repoChanges <- RepoChange{Repo: repo}:
				log.Debugf("posted repo")
			case <-ctx.Done():
				return
			}
			change = noChange
		case <-s.manualRefresh:
			// A refresh repo request, trigger repo change immediately
			log.Infof("refresh repo request")
			select {
			case s.repoChanges <- RepoChange{IsStarting: true}:
			case <-ctx.Done():
				return
			}
			wait = time.After(batchInterval)
			change = noChange
			s.triggerRepo(ctx, true)

		case <-ctx.Done():
			// Closing this repo
			return
		}
	}
}

func (s *repoService) triggerStatus(ctx context.Context, repo Repo) {
	go func() {
		repo, err := s.getFreshStatus(repo)
		if err != nil {
			select {
			case s.repoChanges <- RepoChange{Error: err}:
				log.Infof("posted status error")
			case <-ctx.Done():
				return
			}
			return
		}
		s.internalPostRepo(repo)
	}()
}

func (s *repoService) triggerRepo(ctx context.Context, isTriggerFetch bool) {
	log.Infof("TriggerRefreshRepo")
	go func() {
		repo, err := s.GetFreshRepo()
		if err != nil {
			select {
			case s.repoChanges <- RepoChange{Error: err}:
				log.Infof("posted repo error")
			case <-ctx.Done():
				return
			}
			return
		}
		s.internalPostRepo(repo)
		if isTriggerFetch {
			go func() {
				if err := s.gitFetch(); err != nil {
					log.Warnf("Failed to fetch %v", err)
				}
			}()
		}
	}()
}

func (s *repoService) gitFetch() error {
	// pull meta data, but ignore error, if error is key not exist, it can be ignored,
	// if error is remote error, the fetch will handle that
	_ = s.pullMetaData()

	return s.git.Fetch()
}

func (s *repoService) internalPostRepo(repo Repo) {
	select {
	case s.repo <- repo:
		log.Debugf("Post repo")
	default:
		log.Debugf("repo channel done")
	}
}

func (s *repoService) getFreshStatus(repo Repo) (Repo, error) {
	t := time.Now()
	gitStatus, err := s.git.GetStatus()
	if err != nil {
		return Repo{}, err
	}
	repo.Status = newStatus(gitStatus)
	log.Infof("Git status %v", time.Since(t))
	return repo, nil
}

func (s *repoService) GetFreshRepo() (Repo, error) {
	log.Infof("Getting fresh repo for %s", s.git.RepoPath())
	st := timer.Start()
	repo := newRepo()
	repo.RepoPath = s.git.RepoPath()

	gitRepo, err := s.getGitRepo(partialMax)
	if err != nil {
		return Repo{}, err
	}
	repo.MetaData = gitRepo.MetaData

	repo.Status = newStatus(gitRepo.Status)
	repo.Tags = toTags(gitRepo.Tags)
	repo.setGitBranches(gitRepo.Branches)
	repo.setGitCommits(gitRepo.Commits)

	branchesChildren := repo.MetaData.BranchesChildren
	s.branchesService.setBranchForAllCommits(repo, branchesChildren)
	log.Infof("Repo %v: %d commits, %d branches, %d tags, status: %q (%q)", st, len(gitRepo.Commits), len(gitRepo.Branches), len(gitRepo.Tags), &gitRepo.Status, gitRepo.RootPath)
	return *repo, nil
}

func (s *repoService) fetchRoutine(ctx context.Context) {
	fetchTicker := time.NewTicker(fetchInterval)
	go func() {
		<-ctx.Done()
		fetchTicker.Stop()
	}()

	for range fetchTicker.C {
		if err := s.gitFetch(); err != nil {
			log.Warnf("Failed to fetch %v", err)
		}
	}
}

func (t *repoService) getGitRepo(maxCommitCount int) (gitRepo, error) {
	commits, err := t.git.GetLogMax(maxCommitCount)
	if err != nil {
		return gitRepo{}, err
	}
	branches, err := t.git.GetBranches()
	if err != nil {
		return gitRepo{}, err
	}
	status, err := t.git.GetStatus()
	if err != nil {
		return gitRepo{}, err
	}
	tags, err := t.git.GetTags()
	if err != nil {
		return gitRepo{}, err
	}
	metaData := t.getMetaData()

	return gitRepo{
		RootPath: t.git.RepoPath(),
		Commits:  commits,
		Branches: branches,
		Status:   status,
		Tags:     tags,
		MetaData: metaData,
	}, nil
}

func (t *repoService) getMetaData() MetaData {
	metaDataText, err := t.git.GetKeyValue(metaDataKey)
	if err != nil {
		// metaData might not exit yet, use empty string and set local value
		_ = t.git.SetKeyValue(metaDataKey, "")
		metaDataText = ""
	}

	return toMetaData(metaDataText)
}

func (t *repoService) setMetaData(metaData MetaData) error {
	metaDataJson := string(utils.MustJsonMarshal(metaData))
	err := t.git.SetKeyValue(metaDataKey, metaDataJson)
	if err != nil {
		return err
	}

	return t.pushMetaData()
}

func (t *repoService) pullMetaData() error {
	err := t.git.PullKeyValue(metaDataKey)

	if err != nil {
		// Could not pull value, if the reason is key does not exist, then lets set it
		if strings.Contains(err.Error(), "couldn't find remote ref") {
			_, err = t.git.GetKeyValue(metaDataKey)
			if err != nil {
				// Key not set locally, set default first to have something to push
				_ = t.git.SetKeyValue(metaDataKey, "")
			}

			return t.pushMetaData()
		}
		return err
	}

	return nil
}

func (t *repoService) pushMetaData() error {
	return t.git.PushKeyValue(metaDataKey)
}

func (t *repoService) SetAsParentBranch(b *Branch, name string) error {
	if b.ParentBranch == nil {
		return fmt.Errorf("branch has no parent branch %q", name)
	}
	if !b.ParentBranch.IsAmbiguousBranch {
		return fmt.Errorf("parent branch is not a ambiguous branch %q", b.ParentBranch.Name)
	}

	parentName := b.BaseName()
	otherChildren := lo.Filter(b.ParentBranch.AmbiguousBranches, func(v *Branch, _ int) bool {
		return v.BaseName() != parentName
	})
	otherChildrenNames := lo.Map(otherChildren, func(v *Branch, _ int) string {
		return v.BaseName()
	})

	metaData := t.getMetaData()

	parentChildrenNames, ok := metaData.BranchesChildren[parentName]
	if !ok {
		parentChildrenNames = []string{}
		metaData.BranchesChildren[parentName] = parentChildrenNames
	}

	for _, childName := range otherChildrenNames {
		// Ensure parent branch is not a child of any of the children
		childChildrenNames, ok := metaData.BranchesChildren[childName]
		if ok {
			childChildrenNames = lo.Filter(childChildrenNames, func(v string, _ int) bool {
				return v != parentName
			})
			metaData.BranchesChildren[childName] = childChildrenNames
		}

		// Add child name as child to parent
		if !lo.Contains(parentChildrenNames, childName) {
			parentChildrenNames = append(parentChildrenNames, childName)
			metaData.BranchesChildren[parentName] = parentChildrenNames
		}
	}

	return t.setMetaData(metaData)
}

func (t *repoService) UnsetAsParentBranch(name string) error {
	metaData := t.getMetaData()

	_, ok := metaData.BranchesChildren[name]
	if !ok {
		return nil
	}
	delete(metaData.BranchesChildren, name)

	return t.setMetaData(metaData)
}
