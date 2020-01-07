package gitlib

import (
	"fmt"
	"path"
	"strings"
)

type Status struct {
	Modified     int
	Added        int
	Deleted      int
	Conflicted   int
	IsMerging    bool
	MergeMessage string
}

type statusHandler struct {
	cmd GitCommander
}

func newStatus(cmd GitCommander) *statusHandler {
	return &statusHandler{cmd: cmd}
}

func (s *Status) String() string {
	return fmt.Sprintf("M:%d,A:%d,D:%d,C:%d", s.Modified, s.Added, s.Deleted, s.Conflicted)
}

func (h *statusHandler) getStatus() (Status, error) {
	gitStatus, err := h.cmd.Git("status", "-s", "--porcelain", "--ahead-behind", "--untracked-files=all")
	if err != nil {
		return Status{}, err
	}
	return h.parseStatus(gitStatus)
}

func (h *statusHandler) parseStatus(statusText string) (Status, error) {
	status := Status{}
	lines := strings.Split(statusText, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		if strings.HasPrefix(line, "DD ") ||
			strings.HasPrefix(line, "AU ") ||
			strings.HasPrefix(line, "UA ") {
			// How to reproduce this ???
			status.Conflicted++
		} else if strings.HasPrefix(line, "UU ") {
			status.Conflicted++
		} else if strings.HasPrefix(line, "AA ") {
			status.Conflicted++
		} else if strings.HasPrefix(line, "UD ") {
			status.Conflicted++
		} else if strings.HasPrefix(line, "DU ") {
			status.Conflicted++
		} else if strings.HasPrefix(line, "?? ") || strings.HasPrefix(line, " A ") {
			status.Added++
		} else if strings.HasPrefix(line, " D ") || strings.HasPrefix(line, "D") {
			status.Deleted++
		} else {
			status.Modified++
		}
	}
	status.MergeMessage, status.IsMerging = h.getMergeStatus()
	return status, nil
}

func (h *statusHandler) getMergeStatus() (string, bool) {
	mergeMessage := ""
	//mergeIpPath := path.Join(h.cmd.RepoPath(), ".git", "MERGE_HEAD")
	mergeMsgPath := path.Join(h.cmd.RepoPath(), ".git", "MERGE_MSG")
	msg, err := h.cmd.ReadFile(mergeMsgPath)
	if err != nil {
		return "", false
	}
	mergeMessage = strings.TrimSpace(msg)
	return mergeMessage, true
}
