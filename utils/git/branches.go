package git

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
	"os/exec"
	"strconv"
	"strings"
)

const (
	branchesRegexpText = `(?im)^(\*)?\s+(\(HEAD detached at (\S+)\)|(\S+))\s+(\S+)(\s+)?(\[(\S+)(:\s)?(ahead\s(\d+))?(,\s)?(behind\s(\d+))?(gone)?\])?(\s+)?(.+)?`
)

var branchesRegexp = utils.CompileRegexp(branchesRegexpText)

func getBranches(path string) ([]Branch, error) {
	branchesText, err := getGitBranches(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get git branches, %v", err)
	}
	return parseBranches(branchesText)
}

func getGitBranches(path string) (string, error) {
	cmd := exec.Command("git", "branch", "-vv", "--no-color", "--no-abbrev", "--all")
	cmd.Dir = path

	// Get the git branches output
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git branches, %v", err)
	}
	return string(out), nil
}

func parseBranches(branchesText string) ([]Branch, error) {
	var branches []Branch
	lines := strings.Split(branchesText, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		branch, skip, err := parseBranch(line)
		if err != nil {
			return nil, fmt.Errorf("failed to parse branch line %q, %v", line, err)
		}
		if skip {
			continue
		}
		if branch.IsRemote {
			//todo:skipping remote branches for now
			continue
		}
		branches = append(branches, branch)
	}
	return branches, nil
}

func parseBranch(line string) (Branch, bool, error) {
	match := branchesRegexp.FindStringSubmatch(line)
	if match == nil {
		return Branch{}, true, fmt.Errorf("failed to parse branch line %q", line)
	}
	if isPointBranch(match) {
		return Branch{}, true, nil
	}

	branchName := match[4]
	tipID := match[5]
	isCurrent := match[1] == "*"
	isDetached := strings.TrimSpace(match[3]) != ""
	if isDetached {
		branchName = fmt.Sprintf("(%s)", match[3])
	}
	isRemote := strings.HasPrefix(branchName, "remotes/")
	remoteName := match[8]
	aheadCount, _ := strconv.Atoi(match[11])
	behindCount, _ := strconv.Atoi(match[14])
	isRemoteMissing := match[15] == "gone"
	tipCommitMessage := strings.TrimRight(match[17], "\r")
	id := fmt.Sprintf("%s:local", branchName)
	if isRemote {
		id = fmt.Sprintf("%s:remote", branchName)
	}

	log.Infof("Branch %v", match)
	return Branch{
		ID:               id,
		Name:             branchName,
		TipID:            tipID,
		IsCurrent:        isCurrent,
		IsDetached:       isDetached,
		IsRemote:         isRemote,
		RemoteName:       remoteName,
		AheadCount:       aheadCount,
		BehindCount:      behindCount,
		IsRemoteMissing:  isRemoteMissing,
		TipCommitMessage: tipCommitMessage,
	}, false, nil
}

func isPointBranch(matches []string) bool { return matches[5] == "->" }
