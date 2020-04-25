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
	_, err := t.cmd.Git("fetch", "-f", "--prune", "--tags", "--prune-tags", "origin")
	if err != nil {
		return err
	}
	return nil
}

func (t *remoteService) pushBranch(name string) error {
	refs := fmt.Sprintf("refs/heads/%s:refs/heads/%s", name, name)
	_, err := t.cmd.Git("push", "--porcelain", "origin", "-u", refs)
	if err != nil {
		return err
	}
	return nil
}
