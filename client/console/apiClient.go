package console

// import (
// 	"github.com/michael-reichenauer/gmc/api"
// 	"github.com/michael-reichenauer/gmc/utils/rpc"
// )

// type client struct {
// 	client rpc.ServiceClient
// }

// func NewApiClient(serviceClient rpc.ServiceClient) api.Api {
// 	return &client{client: serviceClient}
// }

// func (t client) GetRecentWorkingDirs(args api.NoArg, rsp *[]string) error {
// 	return t.client.Call(args, rsp)
// }

// func (t client) GetSubDirs(args string, rsp *[]string) error {
// 	return t.client.Call(args, rsp)
// }

// func (t client) OpenRepo(args string, rsp *string) error {
// 	return t.client.Call(args, rsp)
// }

// func (t client) CloseRepo(args string, rsp api.NoRsp) error {
// 	return t.client.Call(args, rsp)
// }

// func (t client) GetRepoChanges(args string, rsp *[]api.RepoChange) error {
// 	return t.client.Call(args, rsp)
// }

// func (t client) TriggerRefreshRepo(args string, rsp api.NoRsp) error {
// 	return t.client.Call(args, rsp)
// }

// func (t client) TriggerSearch(args api.Search, rsp api.NoRsp) error {
// 	return t.client.Call(args, rsp)
// }

// func (t client) GetBranches(args api.GetBranchesReq, rsp *[]api.Branch) error {
// 	return t.client.Call(args, rsp)
// }

// func (t client) GetFiles(args api.FilesReq, rsp *[]string) error {
// 	return t.client.Call(args, rsp)
// }

// func (t *client) GetAmbiguousBranchBranches(args api.AmbiguousBranchBranchesReq, rsp *[]api.Branch) error {
// 	return t.client.Call(args, rsp)
// }

// func (t client) Commit(args api.CommitInfoReq, rsp api.NoRsp) error {
// 	return t.client.Call(args, rsp)
// }

// func (t *client) CleanWorkingFolder(args string, rsp api.NoRsp) error {
// 	return t.client.Call(args, rsp)
// }

// func (t *client) UncommitLastCommit(args string, rsp api.NoRsp) error {
// 	return t.client.Call(args, rsp)
// }

// func (t *client) UndoUncommittedFileChanges(args api.FilesReq, rsp api.NoRsp) error {
// 	return t.client.Call(args, rsp)
// }

// func (t *client) UndoAllUncommittedChanges(args string, rsp api.NoRsp) error {
// 	return t.client.Call(args, rsp)
// }

// func (t *client) UndoCommit(args api.IdReq, rsp api.NoRsp) error {
// 	return t.client.Call(args, rsp)
// }

// func (t client) GetCommitDiff(args api.CommitDiffInfoReq, rsp *api.CommitDiff) error {
// 	return t.client.Call(args, rsp)
// }

// func (t client) GetFileDiff(args api.FileDiffInfoReq, rsp *[]api.CommitDiff) error {
// 	return t.client.Call(args, rsp)
// }

// func (t client) GetCommitDetails(args api.CommitDetailsReq, rsp *api.CommitDetailsRsp) error {
// 	return t.client.Call(args, rsp)
// }

// func (t client) ShowBranch(args api.BranchName, rsp api.NoRsp) error {
// 	return t.client.Call(args, rsp)
// }

// func (t client) HideBranch(args api.BranchName, rsp api.NoRsp) error {
// 	return t.client.Call(args, rsp)
// }

// func (t client) Checkout(args api.CheckoutReq, rsp api.NoRsp) error {
// 	return t.client.Call(args, rsp)
// }

// func (t client) PushBranch(args api.BranchName, rsp api.NoRsp) error {
// 	return t.client.Call(args, rsp)
// }

// func (t client) PullCurrentBranch(args string, rsp api.NoRsp) error {
// 	return t.client.Call(args, rsp)
// }

// func (t client) PullBranch(args api.BranchName, rsp api.NoRsp) error {
// 	return t.client.Call(args, rsp)
// }

// func (t client) MergeBranch(args api.BranchName, rsp api.NoRsp) error {
// 	return t.client.Call(args, rsp)
// }

// func (t client) CreateBranch(args api.BranchName, rsp api.NoRsp) error {
// 	return t.client.Call(args, rsp)
// }

// func (t client) DeleteBranch(args api.BranchName, rsp api.NoRsp) error {
// 	return t.client.Call(args, rsp)
// }

// func (t *client) SetAsParentBranch(args api.SetParentReq, rsp api.NoRsp) error {
// 	return t.client.Call(args, rsp)
// }

// func (t *client) UnsetAsParentBranch(args api.BranchName, rsp api.NoRsp) error {
// 	return t.client.Call(args, rsp)
// }
