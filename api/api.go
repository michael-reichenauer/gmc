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
	GetCommitOpenInBranches(id string) []viewrepo.Branch
	GetCommitOpenOutBranches(id string) []viewrepo.Branch
	CurrentNotShownBranch() (viewrepo.Branch, bool)
	CurrentBranch() (viewrepo.Branch, bool)
	GetLatestBranches(shown bool) []viewrepo.Branch
	GetAllBranches(shown bool) []viewrepo.Branch
	GetShownBranches(master bool) []viewrepo.Branch
	ShowBranch(name string)
	HideBranch(name string)
	SwitchToBranch(name string, name2 string) error
	PushBranch(name string) error
	PullBranch() error
	MergeBranch(name string) error
	CreateBranch(name string) error
	DeleteBranch(name string) error
	GetCommitDiff(id string) (git.CommitDiff, error)
	Commit(message string) error
}
