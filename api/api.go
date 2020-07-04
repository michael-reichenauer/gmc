package api

type Api interface {
	GetRecentWorkingDirs(_ NoArg, dirs *[]string) error
	GetSubDirs(dirPath string, dirs *[]string) error

	OpenRepo(path string, repoID *string) error
	CloseRepo(repoID string, _ NoRsp) error

	GetRepoChanges(repoID string, changes *[]RepoChange) error
	TriggerRefreshRepo(repoID string, _ NoRsp) error
	TriggerSearch(search Search, _ NoRsp) error

	GetBranches(args GetBranches, branches *[]Branch) error
	GetCommitDiff(info CommitDiffInfo, diff *CommitDiff) error

	Commit(info CommitInfo, _ NoRsp) error

	ShowBranch(name BranchName, _ NoRsp) error
	HideBranch(name BranchName, _ NoRsp) error
	Checkout(args Checkout, _ NoRsp) error
	PushBranch(name BranchName, _ NoRsp) error
	PullCurrentBranch(repoID string, _ NoRsp) error
	MergeBranch(name BranchName, _ NoRsp) error
	CreateBranch(name BranchName, _ NoRsp) error
	DeleteBranch(name BranchName, _ NoRsp) error
}
