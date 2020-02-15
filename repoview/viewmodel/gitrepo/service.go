package gitrepo

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"sync"
	"time"
)

//
type Service struct {
	RepoEvents   chan Repo
	StatusEvents chan Status
	ErrorEvents  chan error

	branches *branches

	lock          sync.Mutex
	git           *git.Git
	folderMonitor *monitor
	quit          chan struct{}
}

func ToSid(commitID string) string {
	return git.ToSid(commitID)
}

func NewService() *Service {
	return &Service{
		branches:     newBranches(),
		RepoEvents:   make(chan Repo),
		StatusEvents: make(chan Status),
		ErrorEvents:  make(chan error),
	}
}
func (s *Service) gitLib() *git.Git {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.git
}

func (s *Service) monitor() *monitor {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.folderMonitor
}

func (s *Service) RepoPath() string {
	return s.gitLib().RepoPath()
}

func (s *Service) Close() {
	s.monitor().Close()
	s.lock.Lock()
	defer s.lock.Unlock()
	close(s.quit)
}

func (s *Service) Open(workingFolder string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.quit = make(chan struct{})
	s.git = git.NewGit(workingFolder)
	s.folderMonitor = newMonitor()

	s.folderMonitor.Start(s.git.RepoPath(), s.git.IsIgnored)
	go s.monitorStatusChangesRoutine(s.folderMonitor, s.quit)
	go s.monitorRepoChangesRoutine(s.folderMonitor, s.quit)
	go s.fetchRoutine(s.quit)
}

func (s *Service) TriggerRefreshRepo() {
	go func() {
		repo, err := s.GetFreshRepo()
		if err != nil {
			s.ErrorEvents <- err
			return
		}
		s.RepoEvents <- repo

		s.gitLib().Fetch()
	}()
}

func (s *Service) GetFreshRepo() (Repo, error) {
	gitlib := s.gitLib()
	if gitlib == nil {
		return Repo{}, fmt.Errorf("no repo")
	}
	t := time.Now()
	repo := newRepo()
	repo.RepoPath = gitlib.RepoPath()

	gitCommits, err := gitlib.GetLog()
	if err != nil {
		return Repo{}, err
	}
	gitBranches, err := gitlib.GetBranches()
	if err != nil {
		return Repo{}, err
	}
	gitStatus, err := gitlib.GetStatus()
	if err != nil {
		return Repo{}, err
	}
	repo.Status = newStatus(gitStatus)
	repo.setGitBranches(gitBranches)
	repo.setGitCommits(gitCommits)

	s.branches.setBranchForAllCommits(repo)
	log.Infof("Git repo %v", time.Since(t))
	return *repo, nil
}

func (s *Service) getFreshStatus() (Status, error) {
	t := time.Now()
	gitStatus, err := s.gitLib().GetStatus()
	if err != nil {

		return Status{}, err
	}
	status := newStatus(gitStatus)
	log.Infof("Git status %v", time.Since(t))
	return status, nil
}

func (s *Service) monitorRepoChangesRoutine(monitor *monitor, quit chan struct{}) {
	var ticker *time.Ticker
	tickerChan := func() <-chan time.Time {
		if ticker == nil {
			return nil
		}
		return ticker.C
	}
	for {
		select {
		case <-monitor.RepoChange:
			ticker = time.NewTicker(1 * time.Second)
		case <-tickerChan():
			log.Infof("Detected repo change")
			ticker = nil

			// Repo changed, get new fresh repo and report
			repo, err := s.GetFreshRepo()
			if err != nil {
				s.ErrorEvents <- err
				return
			}
			s.RepoEvents <- repo
		case <-quit:
			return
		}
	}
}

func (s *Service) monitorStatusChangesRoutine(monitor *monitor, quit chan struct{}) {
	var ticker *time.Ticker
	tickerChan := func() <-chan time.Time {
		if ticker == nil {
			return nil
		}
		return ticker.C
	}
	for {
		select {
		case <-monitor.StatusChange:
			ticker = time.NewTicker(1 * time.Second)
		case <-tickerChan():
			log.Infof("Detected status change")
			ticker = nil
			// Status changed, get new fresh status and report
			status, err := s.getFreshStatus()
			if err != nil {
				s.ErrorEvents <- err
				return
			}
			s.StatusEvents <- status
		case <-quit:
			return
		}
	}
}

func (s *Service) fetchRoutine(quit chan struct{}) {
	for {
		select {
		case <-time.After(10 * time.Minute):
			if err := s.gitLib().Fetch(); err != nil {
				log.Warnf("Failed to fetch %v", err)
			}
		case <-quit:
			return
		}
	}
}

func (s *Service) GetCommitDiff(id string) ([]git.FileDiff, error) {
	return s.gitLib().CommitDiff(id)
}
