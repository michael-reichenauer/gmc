package git

import (
	"fmt"
)

// fetch/push from remote origin
type remoteService struct {
	cmd gitCommander
}

func newRemoteService(cmd gitCommander) *remoteService {
	return &remoteService{cmd: cmd}
}

func (t *remoteService) fetch() error {
	// fetch force, prune deleted remote refs, fetch tags, prune deleted tags,
	_, err := t.cmd.Git("fetch", "--force", "--prune", "--tags", "--prune-tags", "origin")
	return err
}

func (t *remoteService) pushBranch(name string) error {
	// push set upstream
	refs := fmt.Sprintf("refs/heads/%s:refs/heads/%s", name, name)
	_, err := t.cmd.Git("push", "--porcelain", "origin", "--set-upstream", refs)
	return err
}

func (t *remoteService) deleteRemoteBranch(name string) error {
	name = StripRemotePrefix(name)
	_, err := t.cmd.Git("push", "--porcelain", "origin", "--delete", name)
	return err
}
