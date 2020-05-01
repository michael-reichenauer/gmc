package gitrepo

import (
	"context"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"time"
)

const (
	fetchInterval = 10 * time.Minute
	batchInterval = 1 * time.Second
)

type RepoChange struct {
	IsStarting bool
	Repo       Repo
	Error      error
}

type GitRepo interface {
	RepoChanges() chan RepoChange
	RepoPath() string
	TriggerManualRefresh()

	StartMonitor(ctx context.Context)

	GetCommitDiff(id string) (git.CommitDiff, error)
	SwitchToBranch(name string) error
	Commit(commit string) error
	PushBranch(name string) error
	CreateBranch(name string) error
	MergeBranch(name string) error
	DeleteRemoteBranch(name string) error
	DeleteLocalBranch(name string) error
	PullBranch() error
}

type gitRepo struct {
	repoChanges chan RepoChange

	branchesService *branchesService
	folderMonitor   *monitor
	git             git.Git
	rootPath        string
	repo            chan Repo
	manualRefresh   chan struct{}
}

func ToSid(commitID string) string {
	return git.ToSid(commitID)
}

func NewGitRepo(workingFolder string) GitRepo {
	g := git.NewGit(workingFolder)
	return &gitRepo{
		rootPath:        workingFolder,
		branchesService: newBranchesService(),
		git:             g,
		folderMonitor:   newMonitor(workingFolder, g),
		repoChanges:     make(chan RepoChange, 1),
		repo:            make(chan Repo, 1),
		manualRefresh:   make(chan struct{}, 1),
	}
}

func (s *gitRepo) RepoChanges() chan RepoChange {
	return s.repoChanges
}

func (s *gitRepo) RepoPath() string {
	return s.git.RepoPath()
}

func (s *gitRepo) StartMonitor(ctx context.Context) {
	go s.monitorRoutine(ctx)
	go s.fetchRoutine(ctx)
}

func (s *gitRepo) GetCommitDiff(id string) (git.CommitDiff, error) {
	log.Infof("Get diff for %q", id)
	return s.git.CommitDiff(id)
}

func (s *gitRepo) SwitchToBranch(name string) error {
	return s.git.Checkout(name)
}

func (s *gitRepo) Commit(message string) error {
	return s.git.Commit(message)
}

func (s *gitRepo) TriggerManualRefresh() {
	select {
	case s.manualRefresh <- struct{}{}:
	default:
		log.Infof("TriggerManualRefresh full")
	}
}

func (s *gitRepo) monitorRoutine(ctx context.Context) {
	log.Infof("monitorRoutine start")
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
			// This is to avoid that multiple change events in a short interval is batched
			if change != noChange {
				log.Infof("waited for %v", change)
				if change == statusChange && hasRepo {
					s.triggerStatus(repo)
				} else {
					s.triggerRepo(false)
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
				// This is to avoid that multiple change events in a short interval is batched
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
			log.Infof("Received repo")
			hasRepo = true
			select {
			case s.repoChanges <- RepoChange{Repo: repo}:
				log.Infof("posted repo")
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
			s.triggerRepo(true)

		case <-ctx.Done():
			// Closing this repo
			return
		}
	}
}

func (s *gitRepo) triggerStatus(repo Repo) {
	go func() {
		repo, err := s.getFreshStatus(repo)
		if err != nil {
			return
		}
		s.internalPostRepo(repo)
	}()
}

func (s *gitRepo) triggerRepo(isTriggerFetch bool) {
	log.Infof("TriggerRefreshRepo")
	go func() {
		repo, err := s.getFreshRepo()
		if err != nil {
			return
		}
		s.internalPostRepo(repo)
		if isTriggerFetch {
			go func() {
				if err := s.git.Fetch(); err != nil {
					log.Warnf("Failed to fetch %v", err)
				}
			}()
		}
	}()
}

func (s *gitRepo) internalPostRepo(repo Repo) {
	select {
	case s.repo <- repo:
		log.Infof("Post repo")
	default:
		log.Infof("repo channel done")
	}
}

func (s *gitRepo) getFreshStatus(repo Repo) (Repo, error) {
	t := time.Now()
	gitStatus, err := s.git.GetStatus()
	if err != nil {
		return Repo{}, err
	}
	repo.Status = newStatus(gitStatus)
	log.Infof("Git status %v", time.Since(t))
	return repo, nil
}

func (s *gitRepo) getFreshRepo() (Repo, error) {
	log.Infof("Getting fresh repo for %s", s.git.RepoPath())
	t := time.Now()
	repo := newRepo()
	repo.RepoPath = s.git.RepoPath()

	gitRepo, err := s.git.GetRepo()
	if err != nil {
		return Repo{}, err
	}

	repo.Status = newStatus(gitRepo.Status)
	repo.Tags = s.toTags(gitRepo.Tags)
	repo.setGitBranches(gitRepo.Branches)
	repo.setGitCommits(gitRepo.Commits)

	s.branchesService.setBranchForAllCommits(repo)
	log.Infof("Git repo %v", time.Since(t))
	return *repo, nil
}

func (s *gitRepo) fetchRoutine(ctx context.Context) {
	fetchTicker := time.NewTicker(fetchInterval)
	go func() {
		<-ctx.Done()
		fetchTicker.Stop()
	}()

	for range fetchTicker.C {
		if err := s.git.Fetch(); err != nil {
			log.Warnf("Failed to fetch %v", err)
		}
	}
}

func (s *gitRepo) PushBranch(name string) error {
	return s.git.PushBranch(name)
}

func (s *gitRepo) PullBranch() error {
	return s.git.PullBranch()
}

func (s *gitRepo) MergeBranch(name string) error {
	return s.git.MergeBranch(name)
}

func (s *gitRepo) CreateBranch(name string) error {
	return s.git.CreateBranch(name)
}

func (s *gitRepo) DeleteRemoteBranch(name string) error {
	return s.git.DeleteRemoteBranch(name)
}

func (s *gitRepo) DeleteLocalBranch(name string) error {
	return s.git.DeleteLocalBranch(name)
}

func (s *gitRepo) toTags(gitTags []git.Tag) []Tag {
	tags := make([]Tag, len(gitTags))
	for i, tag := range gitTags {
		tags[i] = Tag{CommitID: tag.CommitID, TagName: tag.TagName}
	}
	return tags
}
