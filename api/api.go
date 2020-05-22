package api

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
	RepoChanges() chan RepoChange
	GetCommitOpenInBranches(id string) []Branch
	GetCommitOpenOutBranches(id string) []Branch
	CurrentNotShownBranch() (Branch, bool)
	CurrentBranch() (Branch, bool)
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
