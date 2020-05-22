package api

import (
	"github.com/michael-reichenauer/gmc/server/viewrepo"
	"github.com/michael-reichenauer/gmc/utils/git"
)

type Api interface {
	OpenRepo(path string) (Repo, error)
	GetRecentDirs() ([]string, error)
	GetSubDirs(path string) ([]string, error)
}

type Repo interface {
	TriggerRefreshModel()
	TriggerSearch(text string)
	StartMonitor()
	Close()
	RepoChanges() chan viewrepo.RepoChange
	GetCommitOpenInBranches(id string, repo viewrepo.Repo) []viewrepo.Branch
	GetCommitOpenOutBranches(id string, repo viewrepo.Repo) []viewrepo.Branch
	CurrentNotShownBranch(repo viewrepo.Repo) (viewrepo.Branch, bool)
	CurrentBranch(repo viewrepo.Repo) (viewrepo.Branch, bool)
	GetLatestBranches(repo viewrepo.Repo, shown bool) []viewrepo.Branch
	GetAllBranches(repo viewrepo.Repo, shown bool) []viewrepo.Branch
	GetShownBranches(repo viewrepo.Repo, master bool) []viewrepo.Branch
	ShowBranch(name string)
	HideBranch(name string)
	SwitchToBranch(name string, name2 string, repo viewrepo.Repo) error
	PushBranch(name string) error
	PullBranch() error
	MergeBranch(name string) error
	CreateBranch(name string) error
	DeleteBranch(name string, repo viewrepo.Repo) error
	GetCommitDiff(id string) (git.CommitDiff, error)
	Commit(message string) error
}
