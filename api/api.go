package api

type Api interface {
	GetRecentWorkingDirs(_ Nil, dirs *[]string) error
	GetSubDirs(dirPath string, dirs *[]string) error

	OpenRepo(path string, _ Nil) error
	CloseRepo(_ Nil, _ Nil) error
	GetChanges(_ Nil, changes *[]RepoChange) error

	TriggerRefreshModel(_ Nil, _ Nil) error
	TriggerSearch(text string, _ Nil) error
	GetCommitOpenInBranches(id string, branches *[]Branch) error
	GetCommitOpenOutBranches(id string, branches *[]Branch) error
	GetCurrentNotShownBranch(_ Nil, branch *Branch) error
	GetCurrentBranch(_ Nil, branch *Branch) error
	GetLatestBranches(shown bool, branches *[]Branch) error
	GetAllBranches(shown bool, branches *[]Branch) error
	GetShownBranches(master bool, branches *[]Branch) error
	ShowBranch(name string, _ Nil) error
	HideBranch(name string, _ Nil) error
	SwitchToBranch(args SwitchArgs, _ Nil) error
	PushBranch(name string, _ Nil) error
	PullBranch(_ Nil, _ Nil) error
	MergeBranch(name string, _ Nil) error
	CreateBranch(name string, _ Nil) error
	DeleteBranch(name string, _ Nil) error
	GetCommitDiff(id string, diff *CommitDiff) error
	Commit(message string, _ Nil) error
}
