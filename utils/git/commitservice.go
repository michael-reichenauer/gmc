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

func (t *commitService) commitAllChanges(message string) error {
	// Encode '"' chars
	message = strings.ReplaceAll(message, "\"", "\\\"")
	if !t.isMergeInProgress() {
		_, err := t.cmd.Git("add", ".")
		if err != nil {
			return fmt.Errorf("failed to stage before commit, %v", err)
		}
	}

	_, err := t.cmd.Git("commit", "-am", message)
	if err != nil {
		return fmt.Errorf("failed to commit, %v", err)
	}
	return nil
}

func (t *commitService) isMergeInProgress() bool {
	mergeHeadPath := path.Join(t.cmd.RepoPath(), ".git", "MERGE_HEAD")
	return utils.FileExists(mergeHeadPath)
}
