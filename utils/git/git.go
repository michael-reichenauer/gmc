package git

import (
	"os/exec"
	"strings"
)

const (
	UncommittedID  = "0000000000000000000000000000000000000000"
	UncommittedSID = "000000"
)

type Git struct {
	cmd           GitCommander
	status        *statusHandler
	logHandler    *logHandler
	branches      *branchesHandler
	fetchService  *fetchService
	ignoreHandler *ignoreHandler
	diffService   *diffService
	commitService *commitService
	pushService   *pushService
}

func NewGit(path string) *Git {
	cmd := newGitCmd(path)
	status := newStatus(cmd)
	return &Git{
		cmd:           cmd,
		status:        status,
		logHandler:    newLog(cmd),
		branches:      newBranches(cmd),
		fetchService:  newFetch(cmd),
		ignoreHandler: newIgnoreHandler(path),
		diffService:   newDiff(cmd, status),
		commitService: newCommit(cmd),
		pushService:   newPush(cmd),
	}
}
func (h *Git) RepoPath() string {
	return h.cmd.RepoPath()
}

func (h *Git) GetLog() ([]Commit, error) {
	return h.logHandler.getLog()
}

func (h *Git) GetBranches() ([]Branch, error) {
	return h.branches.getBranches()
}

func (h *Git) GetStatus() (Status, error) {
	return h.status.getStatus()
}

func (h *Git) Fetch() error {
	return h.fetchService.fetch()
}

func (h *Git) CommitDiff(id string) (CommitDiff, error) {
	return h.diffService.commitDiff(id)
}

func (h *Git) IsIgnored(path string) bool {
	return h.ignoreHandler.isIgnored(path)
}

func (h *Git) Checkout(name string) error {
	return h.branches.checkout(name)
}

func (h *Git) Commit(message string) error {
	return h.commitService.commitAllChanges(message)
}

func (h *Git) PushBranch(name string) error {
	return h.pushService.pushBranch(name)
}

func (h *Git) MergeBranch(name string) error {
	return h.branches.mergeBranch(name)
}

func (h *Git) CreateBranch(name string) error {
	return h.branches.createBranch(name)
}

// GitVersion returns the git version
func Version() string {
	out, _ := exec.Command("git", "version").Output()
	return strings.TrimSpace(string(out))
}
