package api

import "github.com/michael-reichenauer/gmc/utils/async"

type Api interface {
	GetRecentWorkingDirs() ([]string, error)
	GetSubDirs(dirPath string) ([]string, error)

	OpenRepo(path string) async.Task[string]
	CloseRepo(repoID string) error

	GetRepoChanges(repoID string) ([]RepoChange, error)
	TriggerRefreshRepo(repoID string) error
	TriggerSearch(search Search) error

	GetBranches(args GetBranchesReq) ([]Branch, error)
	GetFiles(args FilesReq) ([]string, error)
	GetCommitDiff(info CommitDiffInfoReq) (CommitDiff, error)
	GetFileDiff(info FileDiffInfoReq) ([]CommitDiff, error)
	GetCommitDetails(req CommitDetailsReq) (CommitDetailsRsp, error)
	GetAmbiguousBranchBranches(args AmbiguousBranchBranchesReq, branches *[]Branch) error

	Commit(info CommitInfoReq, _ NoRsp) error
	UndoCommit(id IdReq, _ NoRsp) error
	UndoUncommittedFileChanges(info FilesReq, _ NoRsp) error
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
