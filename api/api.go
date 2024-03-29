package api

import "github.com/michael-reichenauer/gmc/utils/async"

type Api interface {
	GetRecentWorkingDirs() ([]string, error)
	GetSubDirs(dirPath string) ([]string, error)

	OpenRepo(path string) async.Task[string]
	CloneRepo(uri, path string) async.Task[any]
	CloseRepo(repoID string) error

	GetRepoChanges(repoID string) ([]RepoChange, error)
	TriggerRefreshRepo(repoID string) error
	TriggerSearch(search Search) error

	GetBranches(args GetBranchesReq) ([]Branch, error)
	GetFiles(args FilesReq) ([]string, error)
	GetCommitDiff(info CommitDiffInfoReq) (CommitDiff, error)
	GetFileDiff(info FileDiffInfoReq) ([]CommitDiff, error)
	GetCommitDetails(req CommitDetailsReq) (CommitDetailsRsp, error)
	GetAmbiguousBranchBranches(args AmbiguousBranchBranchesReq) ([]Branch, error)

	Commit(info CommitInfoReq) error
	UndoCommit(repoID, id string) error
	UndoUncommittedFileChanges(repoID, path string) error
	UncommitLastCommit(repoID string) error
	UndoAllUncommittedChanges(repoID string) error
	CleanWorkingFolder(repoID string) error

	ShowBranch(name BranchName) error
	HideBranch(name BranchName) error

	Checkout(repoId, name, displayName string) error
	PushBranch(repoID, branchName string) error
	PullCurrentBranch(repoID string) error
	PullBranch(name BranchName) error
	MergeBranch(name BranchName) error
	MergeSquashBranch(repoID, branchName string) error
	CreateBranch(name BranchName) error
	DeleteBranch(repoID, branchName string, isForced bool) error
	SetAsParentBranch(req SetParentReq) error
	UnsetAsParentBranch(name BranchName) error
}
