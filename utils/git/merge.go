package git

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils/log"
	"strings"
)

// fetches from remote origin
type mergeService struct {
	cmd GitCommander
}

func newMerge(cmd GitCommander) *mergeService {
	return &mergeService{cmd: cmd}
}

func (h *mergeService) mergeBranch(name string) error {
	// $"merge --no-ff --no-commit --stat --progress {name}", ct);
	output, err := h.cmd.Git("merge", "--no-ff", "--no-commit", "--stat", name)
	if err != nil {
		log.Infof("output %q", output)
		if strings.Contains(err.Error(), "exit status 1") &&
			strings.Contains(output, "CONFLICT (\"") {
			return fmt.Errorf("merge of %s resulted in conflict(s)", name)
		}
		return err
	}
	return nil
}
