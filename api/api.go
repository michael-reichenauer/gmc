package api

type Api interface {
	GetRecentWorkingDirs(_ NoArg, dirs *[]string) error
	GetSubDirs(dirPath string, dirs *[]string) error

	OpenRepo(path string, _ NoRsp) error
	CloseRepo(_ NoArg, _ NoRsp) error

	GetRepoChanges(id string, changes *[]RepoChange) error
	TriggerRefreshRepo(_ NoArg, _ NoRsp) error
	TriggerSearch(text string, _ NoRsp) error

	GetBranches(args GetBranchesArgs, branches *[]Branch) error
	GetCommitDiff(id string, diff *CommitDiff) error

	Commit(message string, _ NoRsp) error

	ShowBranch(name string, _ NoRsp) error
	HideBranch(name string, _ NoRsp) error
	Checkout(args CheckoutArgs, _ NoRsp) error
	PushBranch(name string, _ NoRsp) error
	PullCurrentBranch(_ NoArg, _ NoRsp) error
	MergeBranch(name string, _ NoRsp) error
	CreateBranch(name string, _ NoRsp) error
	DeleteBranch(name string, _ NoRsp) error
}
