package git

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
	AddedFiles   []string
}

type statusService struct {
	cmd gitCommander
}

func newStatus(cmd gitCommander) *statusService {
	return &statusService{cmd: cmd}
}

func (t *Status) String() string {
	return fmt.Sprintf("M:%d,A:%d,D:%d,C:%d", t.Modified, t.Added, t.Deleted, t.Conflicted)
}

func (t *statusService) getStatus() (Status, error) {
	gitStatus, err := t.cmd.Git("status", "-s", "--porcelain", "--ahead-behind", "--untracked-files=all")
	if err != nil {
		return Status{}, err
	}
	return t.parseStatus(gitStatus)
}

func (t *statusService) parseStatus(statusText string) (Status, error) {
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
			status.AddedFiles = append(status.AddedFiles, line[3:])
		} else if strings.HasPrefix(line, " D ") || strings.HasPrefix(line, "D") {
			status.Deleted++
		} else {
			status.Modified++
		}
	}
	status.MergeMessage, status.IsMerging = t.getMergeStatus()
	return status, nil
}

func (t *statusService) getMergeStatus() (string, bool) {
	mergeMessage := ""
	//mergeIpPath := path.Join(h.cmd.RepoPath(), ".git", "MERGE_HEAD")
	mergeMsgPath := path.Join(t.cmd.RepoPath(), ".git", "MERGE_MSG")
	msg, err := t.cmd.ReadFile(mergeMsgPath)
	if err != nil {
		return "", false
	}
	lines := strings.Split(msg, "\n")
	mergeMessage = strings.TrimSpace(lines[0])
	return mergeMessage, true
}
