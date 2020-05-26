package api

type Api interface {
	GetRecentWorkingDirs(_ NoArg, dirs *[]string) error
	GetSubDirs(dirPath string, dirs *[]string) error

	OpenRepo(path string, _ NoRsp) error
	CloseRepo(_ NoArg, _ NoRsp) error
	GetChanges(_ NoArg, changes *[]RepoChange) error

	TriggerRefreshModel(_ NoArg, _ NoRsp) error
	TriggerSearch(text string, _ NoRsp) error
	GetCommitOpenInBranches(id string, branches *[]Branch) error
	GetCommitOpenOutBranches(id string, branches *[]Branch) error
	GetCurrentNotShownBranch(_ NoArg, branch *Branch) error
	GetCurrentBranch(_ NoArg, branch *Branch) error
	GetLatestBranches(shown bool, branches *[]Branch) error
	GetAllBranches(shown bool, branches *[]Branch) error
	GetShownBranches(master bool, branches *[]Branch) error
	ShowBranch(name string, _ NoRsp) error
	HideBranch(name string, _ NoRsp) error
	SwitchToBranch(args SwitchArgs, _ NoRsp) error
	PushBranch(name string, _ NoRsp) error
	PullBranch(_ NoArg, _ NoRsp) error
	MergeBranch(name string, _ NoRsp) error
	CreateBranch(name string, _ NoRsp) error
	DeleteBranch(name string, _ NoRsp) error
	GetCommitDiff(id string, diff *CommitDiff) error
	Commit(message string, _ NoRsp) error
}
