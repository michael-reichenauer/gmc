package git

import (
	"fmt"
)

// fetches from remote origin
type pushService struct {
	cmd GitCommander
}

func newPush(cmd GitCommander) *pushService {
	return &pushService{cmd: cmd}
}

func (h *pushService) pushBranch(name string) error {
	refs := fmt.Sprintf("refs/heads/%s:refs/heads/%s", name, name)
	_, err := h.cmd.Git("push", "--porcelain", "origin", "-u", refs)
	if err != nil {
		return err
	}
	return nil
}
