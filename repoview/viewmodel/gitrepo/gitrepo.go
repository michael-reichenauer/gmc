package gitrepo

import (
	"github.com/michael-reichenauer/gmc/utils/gitlib"
	"github.com/michael-reichenauer/gmc/utils/log"
	"time"
)

type Handler struct {
	RepoEvents   chan Repo
	StatusEvents chan Status
	ErrorEvents  chan error

	gitLib   *gitlib.Repo
	monitor  *monitor
	branches *branches
}

func NewModel(repoPath string) *Handler {
	gitLib := gitlib.NewRepo(repoPath)
	return &Handler{
		gitLib:       gitLib,
		monitor:      newMonitor(gitLib.RepoPath()),
		branches:     newBranches(),
		RepoEvents:   make(chan Repo),
		StatusEvents: make(chan Status),
		ErrorEvents:  make(chan error),
	}
}

func (h *Handler) StartRepoMonitor() {
	h.monitor.Start()
	go h.monitorStatusChangesRoutine()
	go h.monitorRepoChangesRoutine()
}

func (h *Handler) TriggerRefreshRepo() {
	go func() {
		repo, err := h.GetFreshRepo()
		if err != nil {
			h.ErrorEvents <- err
			return
		}
		h.RepoEvents <- repo
	}()
}

func (h *Handler) GetFreshRepo() (Repo, error) {
	t := time.Now()
	repo := newRepo()
	repo.RepoPath = h.gitLib.RepoPath()

	gitCommits, err := h.gitLib.GetLog()
	if err != nil {
		return Repo{}, err
	}
	gitBranches, err := h.gitLib.GetBranches()
	if err != nil {
		return Repo{}, err
	}
	gitStatus, err := h.gitLib.GetStatus()
	if err != nil {
		return Repo{}, err
	}
	repo.Status = newStatus(gitStatus)
	repo.setGitBranches(gitBranches)
	repo.setGitCommits(gitCommits)

	h.branches.setBranchForAllCommits(repo)
	log.Infof("Git repo %v", time.Since(t))
	return *repo, nil
}

func (h *Handler) getFreshStatus() (Status, error) {
	t := time.Now()
	gitStatus, err := h.gitLib.GetStatus()
	if err != nil {

		return Status{}, err
	}
	status := newStatus(gitStatus)
	log.Infof("Git status %v", time.Since(t))
	return status, nil
}

func (h *Handler) monitorRepoChangesRoutine() {
	var ticker *time.Ticker
	tickerChan := func() <-chan time.Time {
		if ticker == nil {
			return nil
		}
		return ticker.C
	}
	for {
		select {
		case <-h.monitor.RepoChange:
			ticker = time.NewTicker(1 * time.Second)
		case <-tickerChan():
			log.Infof("Detected repo change")
			ticker = nil

			// Repo changed, get new fresh repo and report
			repo, err := h.GetFreshRepo()
			if err != nil {
				h.ErrorEvents <- err
				return
			}
			h.RepoEvents <- repo
		}
	}
}

func (h *Handler) monitorStatusChangesRoutine() {
	var ticker *time.Ticker
	tickerChan := func() <-chan time.Time {
		if ticker == nil {
			return nil
		}
		return ticker.C
	}
	for {
		select {
		case <-h.monitor.StatusChange:
			ticker = time.NewTicker(1 * time.Second)
		case <-tickerChan():
			log.Infof("Detected status change")
			ticker = nil
			// Status changed, get new fresh status and report
			status, err := h.getFreshStatus()
			if err != nil {
				h.ErrorEvents <- err
				return
			}
			h.StatusEvents <- status
		}
	}
}
