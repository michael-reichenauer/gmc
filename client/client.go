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

func (t client) OpenRepo(path string) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.OpenRepo(path)
}

func (t client) GetRecentWorkingDirs() ([]string, error) {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.GetRecentWorkingDirs()
}

func (t client) GetSubDirs(path string) ([]string, error) {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.GetSubDirs(path)
}

func (t client) CloseRepo() {
	log.Debugf(">")
	defer log.Debugf("<")
	t.server.CloseRepo()
}

func (t client) GetChanges() []api.RepoChange {
	return t.server.GetChanges()
}

func (t client) TriggerRefreshModel() {
	log.Debugf(">")
	defer log.Debugf("<")
	t.server.TriggerRefreshModel()
}

func (t client) TriggerSearch(text string) {
	log.Debugf(">")
	defer log.Debugf("<")
	t.server.TriggerSearch(text)
}

func (t client) GetCommitOpenInBranches(id string) []api.Branch {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.GetCommitOpenInBranches(id)
}

func (t client) GetCommitOpenOutBranches(id string) []api.Branch {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.GetCommitOpenOutBranches(id)
}

func (t client) GetCurrentNotShownBranch() (api.Branch, bool) {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.GetCurrentNotShownBranch()
}

func (t client) GetCurrentBranch() (api.Branch, bool) {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.GetCurrentBranch()
}

func (t client) GetLatestBranches(shown bool) []api.Branch {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.GetLatestBranches(shown)
}

func (t client) GetAllBranches(shown bool) []api.Branch {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.GetAllBranches(shown)
}

func (t client) GetShownBranches(master bool) []api.Branch {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.GetShownBranches(master)
}

func (t client) ShowBranch(name string) {
	log.Debugf(">")
	defer log.Debugf("<")
	t.server.ShowBranch(name)
}

func (t client) HideBranch(name string) {
	log.Debugf(">")
	defer log.Debugf("<")
	t.server.HideBranch(name)
}

func (t client) SwitchToBranch(name string, name2 string) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.SwitchToBranch(name, name2)
}

func (t client) PushBranch(name string) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.PushBranch(name)
}

func (t client) PullBranch() error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.PullBranch()
}

func (t client) MergeBranch(name string) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.MergeBranch(name)
}

func (t client) CreateBranch(name string) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.CreateBranch(name)
}

func (t client) DeleteBranch(name string) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.DeleteBranch(name)
}

func (t client) GetCommitDiff(id string) (api.CommitDiff, error) {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.GetCommitDiff(id)
}

func (t client) Commit(message string) error {
	log.Debugf(">")
	defer log.Debugf("<")
	return t.server.Commit(message)
}
