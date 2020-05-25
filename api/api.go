package api

type Api interface {
	GetRecentWorkingDirs(_ None, dirs *[]string) error
	GetSubDirs(dirPath string, dirs *[]string) error

	OpenRepo(path string, _ None) error
	CloseRepo(_ None, _ None) error
	GetChanges(_ None, changes *[]RepoChange) error

	TriggerRefreshModel(_ None, _ None) error
	TriggerSearch(text string, _ None) error
	GetCommitOpenInBranches(id string, branches *[]Branch) error
	GetCommitOpenOutBranches(id string, branches *[]Branch) error
	GetCurrentNotShownBranch(_ None, branch *Branch) error
	GetCurrentBranch(_ None, branch *Branch) error
	GetLatestBranches(shown bool, branches *[]Branch) error
	GetAllBranches(shown bool, branches *[]Branch) error
	GetShownBranches(master bool, branches *[]Branch) error
	ShowBranch(name string, _ None) error
	HideBranch(name string, _ None) error
	SwitchToBranch(args SwitchArgs, _ None) error
	PushBranch(name string, _ None) error
	PullBranch(_ None, _ None) error
	MergeBranch(name string, _ None) error
	CreateBranch(name string, _ None) error
	DeleteBranch(name string, _ None) error
	GetCommitDiff(id string, diff *CommitDiff) error
	Commit(message string, _ None) error
}
