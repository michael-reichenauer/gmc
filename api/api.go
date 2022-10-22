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
	GetFiles(args FilesReq, files *[]string) error
	GetCommitDiff(info CommitDiffInfoReq, diff *CommitDiff) error
	GetFileDiff(info FileDiffInfoReq, diff *[]CommitDiff) error
	GetCommitDetails(req CommitDetailsReq, rsp *CommitDetailsRsp) error
	GetAmbiguousBranchBranches(args AmbiguousBranchBranchesReq, branches *[]Branch) error

	Commit(info CommitInfoReq, _ NoRsp) error
	UndoCommit(id IdReq, _ NoRsp) error
	UncommitLastCommit(repoID string, _ NoRsp) error
	UndoAllUncommittedChanges(repoID string, _ NoRsp) error
	CleanWorkingFolder(repoID string, _ NoRsp) error

	ShowBranch(name BranchName, _ NoRsp) error
	HideBranch(name BranchName, _ NoRsp) error

	Checkout(args CheckoutReq, _ NoRsp) error
	PushBranch(name BranchName, _ NoRsp) error
	PullCurrentBranch(repoID string, _ NoRsp) error
	PullBranch(name BranchName, _ NoRsp) error
	MergeBranch(name BranchName, _ NoRsp) error
	CreateBranch(name BranchName, _ NoRsp) error
	DeleteBranch(name BranchName, _ NoRsp) error
	SetAsParentBranch(req SetParentReq, _ NoRsp) error
	UnsetAsParentBranch(name BranchName, _ NoRsp) error
}
