package git

import (
	"fmt"
	"path"
	"strings"

	"github.com/michael-reichenauer/gmc/utils"
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
	mergeHeadPath := path.Join(t.cmd.WorkingDir(), ".git", "MERGE_HEAD")
	return utils.FileExists(mergeHeadPath)
}

func (t *commitService) undoAllUncommittedChanges() error {
	_, err := t.cmd.Git("reset", "--hard")
	if err != nil {
		return fmt.Errorf("failed to reset, %v", err)
	}

	_, err = t.cmd.Git("clean", "-fd")
	if err != nil {
		return fmt.Errorf("failed to clean, %v", err)
	}

	return nil
}

func (t *commitService) cleanWorkingFolder() error {
	_, err := t.cmd.Git("reset", "--hard")
	if err != nil {
		return fmt.Errorf("failed to reset, %v", err)
	}

	_, err = t.cmd.Git("clean", "-fxd")
	if err != nil {
		return fmt.Errorf("failed to clean, %v", err)
	}

	return nil
}

func (t *commitService) undoCommit(id string) error {
	_, err := t.cmd.Git("revert", "--no-commit", id)
	if err != nil {
		return fmt.Errorf("failed to reset, %v", err)
	}

	return nil
}

func (t *commitService) uncommitLastCommit() error {
	_, err := t.cmd.Git("reset", "HEAD~1")
	if err != nil {
		return fmt.Errorf("failed to reset, %v", err)
	}

	return nil
}
