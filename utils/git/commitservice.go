package git

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils"
	"path"
	"strings"
)

// commits
type commitService struct {
	cmd gitCommander
}

func newCommit(cmd gitCommander) *commitService {
	return &commitService{cmd: cmd}
}

func (h *commitService) commitAllChanges(message string) error {
	// Encode '"' chars
	message = strings.ReplaceAll(message, "\"", "\\\"")
	if !h.isMergeInProgress() {
		_, err := h.cmd.Git("add", ".")
		if err != nil {
			return fmt.Errorf("failed to stage before commit, %v", err)
		}
	}

	_, err := h.cmd.Git("commit", "-am", message)
	if err != nil {
		return fmt.Errorf("failed to commit, %v", err)
	}
	return nil
}

func (h *commitService) isMergeInProgress() bool {
	mergeHeadPath := path.Join(h.cmd.RepoPath(), ".git", "MERGE_HEAD")
	return utils.FileExists(mergeHeadPath)
}
