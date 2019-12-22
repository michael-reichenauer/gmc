package gitmodel

import (
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"time"
)

//
type Handler struct {
	RepoEvents   chan Repo
	StatusEvents chan Status
	ErrorEvents  chan error

	gitRepo  *git.Repo
	monitor  *monitor
	branches *branches
}

func NewModel(repoPath string) *Handler {
	gitRepo := git.NewRepo(repoPath)
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
		h.refreshStatus()
	}()
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
			log.Infof("repo event, postpone repo")
			ticker = time.NewTicker(3 * time.Second)
		case <-tickerChan():
			log.Infof("refresh repo")
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
			log.Infof("status event, postpone status")
			ticker = time.NewTicker(3 * time.Second)
		case <-tickerChan():
			log.Infof("refresh status")
			ticker = nil
			h.refreshStatus()
		}
	}
}
func (h *Handler) refreshRepo() {
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

	repo.setGitBranches(gitBranches)
	repo.setGitCommits(gitCommits)

	h.branches.setBranchForAllCommits(repo)
	h.RepoEvents <- *repo
}

func (h *Handler) refreshStatus() {
	gitStatus, err := h.gitRepo.GetStatus()
	if err != nil {
		h.ErrorEvents <- err
		return
	}
	status := newStatus(gitStatus)
	h.StatusEvents <- status
}
