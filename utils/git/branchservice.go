package git

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/michael-reichenauer/gmc/utils"
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

type Branches []Branch

func (t *Branch) String() string {
	return t.Name
}

const (
	branchesRegexpText = `(?im)^(\*)?\s+(\(HEAD detached at (\S+)\)|(\S+))\s+(\S+)(\s+)?(\[(\S+)(:\s)?(ahead\s(\d+))?(,\s)?(behind\s(\d+))?(gone)?\])?(\s+)?(.+)?`
	remotePrefix       = "remotes/"
	originPrefix       = "origin/"
)

var branchesRegexp = utils.CompileRegexp(branchesRegexpText)

type branchesService struct {
	cmd gitCommander
}

func newBranchService(cmd gitCommander) *branchesService {
	return &branchesService{cmd: cmd}
}

func (bs Branches) MustCurrent() Branch {
	for _, b := range bs {
		if b.IsCurrent {
			return b
		}
	}
	panic("no current branch")
}

func (bs Branches) MustByName(name string) Branch {
	for _, b := range bs {
		if name == b.Name {
			return b
		}
	}
	panic("no branch: " + name)
}

func (t *branchesService) checkout(name string) error {
	_, err := t.cmd.Git("checkout", name)
	if err != nil {
		return fmt.Errorf("failed to get checkout %q, %v", name, err)
	}
	return nil
}

func (t *branchesService) getBranches() (Branches, error) {
	branchesText, err := t.cmd.Git("branch", "-vv", "--no-color", "--no-abbrev", "--all")
	if err != nil {
		return nil, fmt.Errorf("failed to get git branches, %v", err)
	}
	return t.parseBranchesOutput(branchesText)
}

func (t *branchesService) mergeBranch(name string) error {
	name = StripRemotePrefix(name)
	// $"merge --no-ff --no-commit --stat --progress {name}", ct);
	output, err := t.cmd.Git("merge", "--no-ff", "--no-commit", "--stat", name)
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") &&
			strings.Contains(output, "CONFLICT") {
			return ErrConflicts
		}
		return err
	}
	return nil
}

func (t *branchesService) mergeSquashBranch(name string) error {
	name = StripRemotePrefix(name)
	// $"merge --no-ff --no-commit --stat --progress {name}", ct);
	output, err := t.cmd.Git("merge", "--no-commit", "--stat", "--squash", name)
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") &&
			strings.Contains(output, "CONFLICT") {
			return ErrConflicts
		}
		return err
	}
	return nil
}

func (t *branchesService) createBranch(name string) error {
	_, err := t.cmd.Git("checkout", "-b", name)
	return err
}

func (t *branchesService) createBranchAt(name string, id string) error {
	_, err := t.cmd.Git("checkout", "-b", name, id)
	return err
}

func (t *branchesService) deleteLocalBranch(name string, isForced bool) error {
	var err error
	if isForced {
		_, err = t.cmd.Git("branch", "--delete", "-D", name)
	} else {
		_, err = t.cmd.Git("branch", "--delete", name)
	}
	return err

}

func (t *branchesService) parseBranchesOutput(branchesText string) ([]Branch, error) {
	var branches []Branch
	lines := strings.Split(branchesText, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		branch, skip, err := t.parseBranchLine(line)
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

func (t *branchesService) parseBranchLine(line string) (Branch, bool, error) {
	match := branchesRegexp.FindStringSubmatch(line)
	if match == nil {
		return Branch{}, true, fmt.Errorf("failed to parse branch line %q", line)
	}
	if t.isPointBranch(match) {
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

func (*branchesService) isPointBranch(matches []string) bool {
	return matches[5] == "->"
}
