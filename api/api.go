package api

type Api interface {
	GetRecentWorkingDirs(_ NoArg, dirs *[]string) error
	GetSubDirs(dirPath string, dirs *[]string) error

	OpenRepo(path string, repoID *string) error
	CloseRepo(repoID string, _ NoRsp) error

	GetRepoChanges(repoID string, changes *[]RepoChange) error
	TriggerRefreshRepo(repoID string, _ NoRsp) error
	TriggerSearch(search Search, _ NoRsp) error

	GetBranches(args GetBranchesReq, branches *[]Branch) error
	GetCommitDiff(info CommitDiffInfoReq, diff *CommitDiff) error
	GetCommitDetails(req CommitDetailsReq, rsp *CommitDetailsRsp) error
	GetAmbiguousBranchBranches(args AmbiguousBranchBranchesReq, branches *[]Branch) error

	Commit(info CommitInfoReq, _ NoRsp) error

	ShowBranch(name BranchName, _ NoRsp) error
	HideBranch(name BranchName, _ NoRsp) error

	Checkout(args CheckoutReq, _ NoRsp) error
	PushBranch(name BranchName, _ NoRsp) error
	PullCurrentBranch(repoID string, _ NoRsp) error
	MergeBranch(name BranchName, _ NoRsp) error
	CreateBranch(name BranchName, _ NoRsp) error
	DeleteBranch(name BranchName, _ NoRsp) error
	SetAsParentBranch(name BranchName, _ NoRsp) error
	UnsetAsParentBranch(name BranchName, _ NoRsp) error
}
