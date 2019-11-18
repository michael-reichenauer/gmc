package git

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
	"io/ioutil"
	"os/exec"
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

func (s *Status) String() string {
	return fmt.Sprintf("M:%d,A:%d,D:%d,C:%d", s.Modified, s.Added, s.Deleted, s.Conflicted)
}

func getStatus(path string) (Status, error) {
	gitStatus, err := getGitStatus(path)
	if err != nil {
		return Status{}, err
	}
	return parseStatus(path, gitStatus)
}

func parseStatus(path, statusText string) (Status, error) {
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
	status.MergeMessage, status.IsMerging = getMergeStatus(path)
	return status, nil
}

func getGitStatus(path string) (string, error) {
	cmd := exec.Command("git", "status", "-s", "--porcelain", "--ahead-behind", "--untracked-files=all")
	cmd.Dir = path

	// Get the git log output
	out, err := cmd.Output()
	if err != nil {
		log.Warnf("Failed %v", err)
		return "", fmt.Errorf("failed to get git log, %v", err)
	}
	return string(out), nil
}

func getMergeStatus(repoPath string) (string, bool) {
	isMergeInProgress := false
	mergeMessage := ""
	mergeIpPath := path.Join(repoPath, ".git", "MERGE_HEAD")
	mergeMsgPath := path.Join(repoPath, ".git", "MERGE_MSG")
	if utils.FileExists(mergeIpPath) {
		isMergeInProgress = true
		msg, _ := ioutil.ReadFile(mergeMsgPath)
		mergeMessage = strings.TrimSpace(string(msg))
	}

	return mergeMessage, isMergeInProgress
}
