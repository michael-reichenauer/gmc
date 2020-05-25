package client

import (
	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/utils/log"
)

type client struct {
	server api.Api
}

func NewClient(server api.Api) api.Api {
	return &client{server: server}
}

func (t client) OpenRepo(args string, rsp api.Nil) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.OpenRepo(args, rsp)
}

func (t client) GetRecentWorkingDirs(args api.Nil, rsp *[]string) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.GetRecentWorkingDirs(args, rsp)
}

func (t client) GetSubDirs(args string, rsp *[]string) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.GetSubDirs(args, rsp)
}

func (t client) CloseRepo(args api.Nil, rsp api.Nil) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.CloseRepo(args, rsp)
}

func (t client) GetChanges(args api.Nil, rsp *[]api.RepoChange) error {
	return t.server.GetChanges(args, rsp)
}

func (t client) TriggerRefreshModel(args api.Nil, rsp api.Nil) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.TriggerRefreshModel(args, rsp)
}

func (t client) TriggerSearch(args string, rsp api.Nil) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.TriggerSearch(args, rsp)
}

func (t client) GetCommitOpenInBranches(args string, rsp *[]api.Branch) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.GetCommitOpenInBranches(args, rsp)
}

func (t client) GetCommitOpenOutBranches(args string, rsp *[]api.Branch) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.GetCommitOpenOutBranches(args, rsp)
}

func (t client) GetCurrentNotShownBranch(args api.Nil, rsp *api.Branch) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.GetCurrentNotShownBranch(args, rsp)
}

func (t client) GetCurrentBranch(args api.Nil, rsp *api.Branch) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.GetCurrentBranch(args, rsp)
}

func (t client) GetLatestBranches(args bool, rsp *[]api.Branch) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.GetLatestBranches(args, rsp)
}

func (t client) GetAllBranches(args bool, rsp *[]api.Branch) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.GetAllBranches(args, rsp)
}

func (t client) GetShownBranches(args bool, rsp *[]api.Branch) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.GetShownBranches(args, rsp)
}

func (t client) ShowBranch(args string, rsp api.Nil) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.ShowBranch(args, rsp)
}

func (t client) HideBranch(args string, rsp api.Nil) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.HideBranch(args, rsp)
}

func (t client) SwitchToBranch(args api.SwitchArgs, rsp api.Nil) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.SwitchToBranch(args, rsp)
}

func (t client) PushBranch(args string, rsp api.Nil) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.PushBranch(args, rsp)
}

func (t client) PullBranch(args api.Nil, rsp api.Nil) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.PullBranch(args, rsp)
}

func (t client) MergeBranch(args string, rsp api.Nil) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.MergeBranch(args, rsp)
}

func (t client) CreateBranch(args string, rsp api.Nil) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.CreateBranch(args, rsp)
}

func (t client) DeleteBranch(args string, rsp api.Nil) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.DeleteBranch(args, rsp)
}

func (t client) GetCommitDiff(args string, rsp *api.CommitDiff) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.GetCommitDiff(args, rsp)
}

func (t client) Commit(args string, rsp api.Nil) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.Commit(args, rsp)
}
