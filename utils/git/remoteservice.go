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

func (t *remoteService) pushRefForce(ref string) error {
	// push set upstream
	refs := fmt.Sprintf("%s:%s", ref, ref)
	_, err := t.cmd.Git("push", "--porcelain", "origin", "--set-upstream", "--force", refs)
	return err
}

func (t *remoteService) pullRef(ref string) error {
	// fetch origin
	refs := fmt.Sprintf("%s:%s", ref, ref)
	_, err := t.cmd.Git("fetch", "origin", refs)
	return err
}

func (t *remoteService) deleteRemoteBranch(name string) error {
	name = StripRemotePrefix(name)
	_, err := t.cmd.Git("push", "--porcelain", "origin", "--delete", name)
	return err
}

func (t *remoteService) pullCurrentBranch() error {
	// "pull --ff --no-rebase --progress"
	_, err := t.cmd.Git("pull", "--ff", "--no-rebase")
	return err
}

func (t *remoteService) pullBranch(name string) error {
	// fetch origin
	branchRefs := fmt.Sprintf("%s:%s", name, name)
	_, err := t.cmd.Git("fetch", "origin", branchRefs)
	return err
}

func (t *remoteService) clone(uri, path string) error {
	// fetch origin
	_, err := t.cmd.Git("clone", uri, path)
	return err
}
