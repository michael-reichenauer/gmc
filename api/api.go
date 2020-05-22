package api

type Api interface {
	OpenRepo(path string) (Repo, error)
	GetRecentDirs() ([]string, error)
	GetSubDirs(path string) ([]string, error)
}

type Repo interface {
	StartMonitor()
	Close()
	RepoChanges() chan RepoChange

	TriggerRefreshModel()
	TriggerSearch(text string)
	GetCommitOpenInBranches(id string) []Branch
	GetCommitOpenOutBranches(id string) []Branch
	GetCurrentNotShownBranch() (Branch, bool)
	GetCurrentBranch() (Branch, bool)
	GetLatestBranches(shown bool) []Branch
	GetAllBranches(shown bool) []Branch
	GetShownBranches(master bool) []Branch
	ShowBranch(name string)
	HideBranch(name string)
	SwitchToBranch(name string, name2 string) error
	PushBranch(name string) error
	PullBranch() error
	MergeBranch(name string) error
	CreateBranch(name string) error
	DeleteBranch(name string) error
	GetCommitDiff(id string) (CommitDiff, error)
	Commit(message string) error
}
