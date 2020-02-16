package git

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils"
	"strconv"
	"strings"
)

type Branch struct {
	Name             string
	DisplayName      string
	TipID            string
	IsCurrent        bool
	IsRemote         bool
	RemoteName       string
	IsDetached       bool
	AheadCount       int
	BehindCount      int
	IsRemoteMissing  bool
	TipCommitMessage string
}

func (b *Branch) String() string {
	return b.Name
}

const (
	branchesRegexpText = `(?im)^(\*)?\s+(\(HEAD detached at (\S+)\)|(\S+))\s+(\S+)(\s+)?(\[(\S+)(:\s)?(ahead\s(\d+))?(,\s)?(behind\s(\d+))?(gone)?\])?(\s+)?(.+)?`
	remotePrefix       = "remotes/"
	originPrefix       = "origin/"
)

var branchesRegexp = utils.CompileRegexp(branchesRegexpText)

type branchesHandler struct {
	cmd GitCommander
}

func newBranches(cmd GitCommander) *branchesHandler {
	return &branchesHandler{cmd: cmd}
}

func (h *branchesHandler) checkout(name string) error {
	_, err := h.cmd.Git("checkout", name)
	if err != nil {
		return fmt.Errorf("failed to get checkout %q, %v", name, err)
	}
	return nil
}

func (h *branchesHandler) getBranches() ([]Branch, error) {
	branchesText, err := h.cmd.Git("branch", "-vv", "--no-color", "--no-abbrev", "--all")
	if err != nil {
		return nil, fmt.Errorf("failed to get git branches, %v", err)
	}
	return h.parseBranchesOutput(branchesText)
}

func (h *branchesHandler) parseBranchesOutput(branchesText string) ([]Branch, error) {
	var branches []Branch
	lines := strings.Split(branchesText, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		branch, skip, err := h.parseBranchLine(line)
		if err != nil {
			return nil, fmt.Errorf("failed to parse branch line %q, %v", line, err)
		}
		if skip {
			continue
		}

		branches = append(branches, branch)
	}
	return branches, nil
}

func (h *branchesHandler) parseBranchLine(line string) (Branch, bool, error) {
	match := branchesRegexp.FindStringSubmatch(line)
	if match == nil {
		return Branch{}, true, fmt.Errorf("failed to parse branch line %q", line)
	}
	if h.isPointBranch(match) {
		return Branch{}, true, nil
	}

	name := match[4]
	isRemote := false
	if strings.HasPrefix(name, remotePrefix) {
		isRemote = true
		name = name[len(remotePrefix):]
	}
	isDetached := strings.TrimSpace(match[3]) != ""
	if isDetached {
		name = fmt.Sprintf("(%s)", match[3])
	}

	displayName := name
	if strings.HasPrefix(name, originPrefix) {
		// make remote branch display name same as local branch name
		displayName = fmt.Sprintf("%s", name[len(originPrefix):])
	}

	tipID := match[5]
	isCurrent := match[1] == "*"

	remoteName := match[8]
	aheadCount, _ := strconv.Atoi(match[11])
	behindCount, _ := strconv.Atoi(match[14])
	isRemoteMissing := match[15] == "gone"
	tipCommitMessage := strings.TrimRight(match[17], "\r")

	return Branch{
		Name:             name,
		DisplayName:      displayName,
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

func (*branchesHandler) isPointBranch(matches []string) bool { return matches[5] == "->" }
