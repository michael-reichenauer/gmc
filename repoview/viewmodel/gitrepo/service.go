package gitrepo

import (
	"github.com/michael-reichenauer/gmc/utils/gitlib"
	"github.com/michael-reichenauer/gmc/utils/log"
	"time"
)

type Service struct {
	RepoEvents   chan Repo
	StatusEvents chan Status
	ErrorEvents  chan error

	gitLib   *gitlib.Repo
	monitor  *monitor
	branches *branches
}

func ToSid(commitID string) string {
	return gitlib.ToSid(commitID)
}

func NewService(repoPath string) *Service {
	gitLib := gitlib.NewRepo(repoPath)
	return &Service{
		gitLib:       gitLib,
		monitor:      newMonitor(gitLib.RepoPath()),
		branches:     newBranches(),
		RepoEvents:   make(chan Repo),
		StatusEvents: make(chan Status),
		ErrorEvents:  make(chan error),
	}
}

func (s *Service) RepoPath() string {
	return s.gitLib.RepoPath()
}

func (s *Service) StartRepoMonitor() {
	s.monitor.Start()
	go s.monitorStatusChangesRoutine()
	go s.monitorRepoChangesRoutine()
	go s.fetchRoutine()
}

func (s *Service) TriggerRefreshRepo() {
	go func() {
		repo, err := s.GetFreshRepo()
		if err != nil {
			s.ErrorEvents <- err
			return
		}
		s.RepoEvents <- repo
		go func() {
			s.gitLib.Fetch()
		}()
	}()
}

func (s *Service) GetFreshRepo() (Repo, error) {
	t := time.Now()
	repo := newRepo()
	repo.RepoPath = s.gitLib.RepoPath()

	gitCommits, err := s.gitLib.GetLog()
	if err != nil {
		return Repo{}, err
	}
	gitBranches, err := s.gitLib.GetBranches()
	if err != nil {
		return Repo{}, err
	}
	gitStatus, err := s.gitLib.GetStatus()
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
	gitStatus, err := s.gitLib.GetStatus()
	if err != nil {

		return Status{}, err
	}
	status := newStatus(gitStatus)
	log.Infof("Git status %v", time.Since(t))
	return status, nil
}

func (s *Service) monitorRepoChangesRoutine() {
	var ticker *time.Ticker
	tickerChan := func() <-chan time.Time {
		if ticker == nil {
			return nil
		}
		return ticker.C
	}
	for {
		select {
		case <-s.monitor.RepoChange:
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
		}
	}
}

func (s *Service) monitorStatusChangesRoutine() {
	var ticker *time.Ticker
	tickerChan := func() <-chan time.Time {
		if ticker == nil {
			return nil
		}
		return ticker.C
	}
	for {
		select {
		case <-s.monitor.StatusChange:
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
		}
	}
}

func (s *Service) fetchRoutine() {
	time.Sleep(300 * time.Millisecond)
	for {
		if err := s.gitLib.Fetch(); err != nil {
			log.Warnf("Failed to fetch %v", err)
		}
		time.Sleep(10 * time.Minute)
	}
}
