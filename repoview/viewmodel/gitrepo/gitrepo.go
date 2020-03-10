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
	Repo  Repo
	Error error
}

type GitRepo struct {
	RepoChanges chan RepoChange

	branchesService *branchesService
	folderMonitor   *monitor
	git             *git.Git
	rootPath        string
	repo            chan Repo
	manualRefresh   chan struct{}
}

func ToSid(commitID string) string {
	return git.ToSid(commitID)
}

func NewGitRepo(workingFolder string) *GitRepo {
	g := git.NewGit(workingFolder)
	return &GitRepo{
		rootPath:        workingFolder,
		branchesService: newBranchesService(),
		git:             g,
		folderMonitor:   newMonitor(workingFolder, g),
		RepoChanges:     make(chan RepoChange),
		repo:            make(chan Repo),
		manualRefresh:   make(chan struct{}),
	}
}

func (s *GitRepo) RepoPath() string {
	return s.git.RepoPath()
}

func (s *GitRepo) StartMonitor(ctx context.Context) {
	go s.monitorRoutine(ctx)
	go s.fetchRoutine(ctx)
}

func (s *GitRepo) GetCommitDiff(id string) (git.CommitDiff, error) {
	log.Infof("Get diff for %q", id)
	return s.git.CommitDiff(id)
}

func (s *GitRepo) SwitchToBranch(name string) {
	s.git.Checkout(name)
}

func (s *GitRepo) TriggerManualRefresh() {
	select {
	case s.manualRefresh <- struct{}{}:
	default:
	}
}

func (s *GitRepo) monitorRoutine(ctx context.Context) {
	log.Infof("monitorRoutine start")
	defer close(s.RepoChanges)
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
				log.Infof("waited for %s", change)
				change = noChange
				if change == statusChange && hasRepo {
					s.triggerStatus(repo)
				} else {
					s.triggerRepo()
				}
			}
			wait = time.After(batchInterval)

		case changeEvent, ok := <-s.folderMonitor.Changes:
			// Received changed repo or status event
			if !ok {
				// No more change events, closing this repo
				return
			}
			log.Infof("Change %s", changeEvent)
			if changeEvent == statusChange && !hasRepo {
				// Status change, but we have not yet a repo, ignore status until repo exist
				log.Infof("No repo for status change %s", changeEvent)
				break
			}
			if change == noChange {
				// First change event, ensure we do wait a while before acting on the event
				// This is to avoid that multiple change events in a short interval is batched
				wait = time.After(batchInterval)
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
			case s.RepoChanges <- RepoChange{Repo: repo}:
				log.Infof("posted repo")
			case <-ctx.Done():
				return
			}
			change = noChange
		case <-s.manualRefresh:
			// A refresh repo request, trigger repo change immediately
			log.Infof("refresh repo request")
			wait = time.After(batchInterval)
			change = noChange
			s.triggerRepo()

		case <-ctx.Done():
			// Closing this repo
			return
		}
	}
}

func (s *GitRepo) triggerStatus(repo Repo) {
	go func() {
		repo, err := s.getFreshStatus(repo)
		if err != nil {
			return
		}
		s.internalPostRepo(repo)
	}()
}

func (s *GitRepo) triggerRepo() {
	log.Infof("TriggerRefreshRepo")
	go func() {
		repo, err := s.getFreshRepo()
		if err != nil {
			return
		}
		s.internalPostRepo(repo)
	}()
}

func (s *GitRepo) internalPostRepo(repo Repo) {
	select {
	case s.repo <- repo:
		log.Infof("Post repo")
	default:
		log.Infof("repo channel done")
	}
}

func (s *GitRepo) TriggerRefreshRepo() {
	log.Infof("TriggerRefreshRepo")
	go func() {
		repo, err := s.getFreshRepo()
		if err != nil {
			return
		}
		s.internalPostRepo(repo)
		if err := s.git.Fetch(); err != nil {
			log.Warnf("Failed to fetch %v", err)
		}
	}()
}

func (s *GitRepo) getFreshStatus(repo Repo) (Repo, error) {
	t := time.Now()
	gitStatus, err := s.git.GetStatus()
	if err != nil {
		return Repo{}, err
	}
	repo.Status = newStatus(gitStatus)
	log.Infof("Git status %v", time.Since(t))
	return repo, nil
}

func (s *GitRepo) getFreshRepo() (Repo, error) {
	log.Infof("Getting fresh repo for %s", s.git.RepoPath())
	t := time.Now()
	repo := newRepo()
	repo.RepoPath = s.git.RepoPath()

	gitCommits, err := s.git.GetLog()
	if err != nil {
		return Repo{}, err
	}
	gitBranches, err := s.git.GetBranches()
	if err != nil {
		return Repo{}, err
	}
	gitStatus, err := s.git.GetStatus()
	if err != nil {
		return Repo{}, err
	}
	repo.Status = newStatus(gitStatus)
	repo.setGitBranches(gitBranches)
	repo.setGitCommits(gitCommits)

	s.branchesService.setBranchForAllCommits(repo)
	log.Infof("Git repo %v", time.Since(t))
	return *repo, nil
}

//
// func (s *GitRepo) monitorRepoChangesRoutine() {
// 	defer s.waitGroup.Done()
// 	var ticker *time.Ticker
// 	tickerChan := func() <-chan time.Time {
// 		if ticker == nil {
// 			return nil
// 		}
// 		return ticker.C
// 	}
// 	for {
// 		select {
// 		case <-s.folderMonitor.RepoChange:
// 			ticker = time.NewTicker(1 * time.Second)
// 		case <-tickerChan():
// 			log.Infof("Detected repo change")
// 			ticker = nil
//
// 			// Repo changed, get new fresh repo and report
// 			repo, err := s.GetFreshRepo()
// 			if err != nil {
// 				s.ErrorEvents <- err
// 				return
// 			}
// 			s.postRepoEvent(repo)
// 		case <-s.done:
// 			return
// 		}
// 	}
// }
//
// func (s *GitRepo) postRepoEvent(repo Repo) {
// 	select {
// 	case s.RepoEvents <- repo:
// 	case <-s.done:
// 	}
// }
//
// func (s *GitRepo) postStatusEvent(status Status) {
// 	select {
// 	case s.StatusEvents <- status:
// 	case <-s.done:
// 	}
// }
//
// func (s *GitRepo) monitorStatusChangesRoutine() {
// 	defer s.waitGroup.Done()
// 	var ticker *time.Ticker
// 	tickerChan := func() <-chan time.Time {
// 		if ticker == nil {
// 			return nil
// 		}
// 		return ticker.C
// 	}
// 	for {
// 		select {
// 		case <-s.StatusEvents:
// 			ticker = time.NewTicker(1 * time.Second)
// 		case <-tickerChan():
// 			log.Infof("Detected status change")
// 			ticker = nil
// 			// Status changed, get new fresh status and report
// 			status, err := s.getFreshStatus()
// 			if err != nil {
// 				s.ErrorEvents <- err
// 				return
// 			}
// 			s.postStatusEvent(status)
// 		case <-s.done:
// 			return
// 		}
// 	}
// }
//
// func (s *GitRepo) getFreshStatus() (Status, error) {
// 	t := time.Now()
// 	gitStatus, err := s.gitLib().GetStatus()
// 	if err != nil {
//
// 		return Status{}, err
// 	}
// 	status := newStatus(gitStatus)
// 	log.Infof("Git status %v", time.Since(t))
// 	return status, nil
// }

func (s *GitRepo) fetchRoutine(ctx context.Context) {
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