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

	gitRepo  *gitlib.Repo
	monitor  *monitor
	branches *branches
}

func NewModel(repoPath string) *Handler {
	gitRepo := gitlib.NewRepo(repoPath)
	return &Handler{
		gitRepo:      gitRepo,
		monitor:      newMonitor(gitRepo.RepoPath),
		branches:     newBranches(),
		RepoEvents:   make(chan Repo),
		StatusEvents: make(chan Status),
		ErrorEvents:  make(chan error),
	}
}

func (h *Handler) Start() {
	h.monitor.Start()
	go h.monitorStatusChangesRoutine()
	go h.monitorRepoChangesRoutine()
}

func (h *Handler) TriggerRefresh() {
	go func() {
		h.refreshRepo()
	}()
}

func (h *Handler) refreshRepo() {
	t := time.Now()
	repo := newRepo()
	repo.RepoPath = h.gitRepo.RepoPath

	gitCommits, err := h.gitRepo.GetLog()
	if err != nil {
		h.ErrorEvents <- err
		return
	}
	gitBranches, err := h.gitRepo.GetBranches()
	if err != nil {
		h.ErrorEvents <- err
		return
	}
	gitStatus, err := h.gitRepo.GetStatus()
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

func (h *Handler) refreshStatus() {
	t := time.Now()
	gitStatus, err := h.gitRepo.GetStatus()
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
			h.refreshRepo()
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
			h.refreshStatus()
		}
	}
}
