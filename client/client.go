package client

import (
	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils/rpc"
)

type client struct {
	client rpc.ServiceClient
}

func NewClient(serviceClient rpc.ServiceClient) api.Api {
	return &client{client: serviceClient}
}

func (t client) OpenRepo(args string, rsp api.NoRsp) error {
	return t.client.Call(args, rsp)
}

func (t client) GetRecentWorkingDirs(args api.NoArg, rsp *[]string) error {
	return t.client.Call(args, rsp)
}

func (t client) GetSubDirs(args string, rsp *[]string) error {
	return t.client.Call(args, rsp)
}

func (t client) CloseRepo(args api.NoArg, rsp api.NoRsp) error {
	return t.client.Call(args, rsp)
}

func (t client) GetRepoChanges(id string, rsp *[]api.RepoChange) error {
	return t.client.Call(id, rsp)
}

func (t client) TriggerRefreshRepo(args api.NoArg, rsp api.NoRsp) error {
	return t.client.Call(args, rsp)
}

func (t client) TriggerSearch(args string, rsp api.NoRsp) error {
	return t.client.Call(args, rsp)
}

func (t client) GetBranches(args api.GetBranchesArgs, rsp *[]api.Branch) error {
	return t.client.Call(args, rsp)
}

func (t client) Commit(args string, rsp api.NoRsp) error {
	return t.client.Call(args, rsp)
}

func (t client) GetCommitDiff(args string, rsp *api.CommitDiff) error {
	return t.client.Call(args, rsp)
}

func (t client) ShowBranch(args string, rsp api.NoRsp) error {
	return t.client.Call(args, rsp)
}

func (t client) HideBranch(args string, rsp api.NoRsp) error {
	return t.client.Call(args, rsp)
}

func (t client) Checkout(args api.CheckoutArgs, rsp api.NoRsp) error {
	return t.client.Call(args, rsp)
}

func (t client) PushBranch(args string, rsp api.NoRsp) error {
	return t.client.Call(args, rsp)
}

func (t client) PullCurrentBranch(args api.NoArg, rsp api.NoRsp) error {
	return t.client.Call(args, rsp)
}

func (t client) MergeBranch(args string, rsp api.NoRsp) error {
	return t.client.Call(args, rsp)
}

func (t client) CreateBranch(args string, rsp api.NoRsp) error {
	return t.client.Call(args, rsp)
}

func (t client) DeleteBranch(args string, rsp api.NoRsp) error {
	return t.client.Call(args, rsp)
}
