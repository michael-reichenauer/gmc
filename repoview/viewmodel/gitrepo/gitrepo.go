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
		monitor:      newMonitor(gitLib.RepoPath),
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

func (h *Handler) TriggerRefresh() {
	go func() {
		h.loadRepo()
	}()
}

func (h *Handler) loadRepo() {
	t := time.Now()
	repo := newRepo()
	repo.RepoPath = h.gitLib.RepoPath

	gitCommits, err := h.gitLib.GetLog()
	if err != nil {
		h.ErrorEvents <- err
		return
	}
	gitBranches, err := h.gitLib.GetBranches()
	if err != nil {
		h.ErrorEvents <- err
		return
	}
	gitStatus, err := h.gitLib.GetStatus()
	if err != nil {
		h.ErrorEvents <- err
		return
	}
	repo.Status = newStatus(gitStatus)
	repo.setGitBranches(gitBranches)
	repo.setGitCommits(gitCommits)

	h.branches.setBranchForAllCommits(repo)
	log.Infof("Git repo %v", time.Since(t))
	h.RepoEvents <- *repo
}

func (h *Handler) loadStatus() {
	t := time.Now()
	gitStatus, err := h.gitLib.GetStatus()
	if err != nil {
		h.ErrorEvents <- err
		return
	}
	status := newStatus(gitStatus)
	log.Infof("Git status %v", time.Since(t))
	h.StatusEvents <- status
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
			h.loadRepo()
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
			h.loadStatus()
		}
	}
}
