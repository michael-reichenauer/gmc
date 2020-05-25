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

func (t client) OpenRepo(args string, rsp api.None) error {
	return t.client.Call(args, rsp)
}

func (t client) GetRecentWorkingDirs(args api.None, rsp *[]string) error {
	return t.client.Call(args, rsp)
}

func (t client) GetSubDirs(args string, rsp *[]string) error {
	return t.client.Call(args, rsp)
}

func (t client) CloseRepo(args api.None, rsp api.None) error {
	return t.client.Call(args, rsp)
}

func (t client) GetChanges(args api.None, rsp *[]api.RepoChange) error {
	return t.client.Call(args, rsp)
}

func (t client) TriggerRefreshModel(args api.None, rsp api.None) error {
	return t.client.Call(args, rsp)
}

func (t client) TriggerSearch(args string, rsp api.None) error {
	return t.client.Call(args, rsp)
}

func (t client) GetCommitOpenInBranches(args string, rsp *[]api.Branch) error {
	return t.client.Call(args, rsp)
}

func (t client) GetCommitOpenOutBranches(args string, rsp *[]api.Branch) error {
	return t.client.Call(args, rsp)
}

func (t client) GetCurrentNotShownBranch(args api.None, rsp *api.Branch) error {
	return t.client.Call(args, rsp)
}

func (t client) GetCurrentBranch(args api.None, rsp *api.Branch) error {
	return t.client.Call(args, rsp)
}

func (t client) GetLatestBranches(args bool, rsp *[]api.Branch) error {
	return t.client.Call(args, rsp)
}

func (t client) GetAllBranches(args bool, rsp *[]api.Branch) error {
	return t.client.Call(args, rsp)
}

func (t client) GetShownBranches(args bool, rsp *[]api.Branch) error {
	return t.client.Call(args, rsp)
}

func (t client) ShowBranch(args string, rsp api.None) error {
	return t.client.Call(args, rsp)
}

func (t client) HideBranch(args string, rsp api.None) error {
	return t.client.Call(args, rsp)
}

func (t client) SwitchToBranch(args api.SwitchArgs, rsp api.None) error {
	return t.client.Call(args, rsp)
}

func (t client) PushBranch(args string, rsp api.None) error {
	return t.client.Call(args, rsp)
}

func (t client) PullBranch(args api.None, rsp api.None) error {
	return t.client.Call(args, rsp)
}

func (t client) MergeBranch(args string, rsp api.None) error {
	return t.client.Call(args, rsp)
}

func (t client) CreateBranch(args string, rsp api.None) error {
	return t.client.Call(args, rsp)
}

func (t client) DeleteBranch(args string, rsp api.None) error {
	return t.client.Call(args, rsp)
}

func (t client) GetCommitDiff(args string, rsp *api.CommitDiff) error {
	return t.client.Call(args, rsp)
}

func (t client) Commit(args string, rsp api.None) error {
	return t.client.Call(args, rsp)
}
