package git

import (
	"os/exec"
	"strings"
)

const (
	UncommittedID  = "0000000000000000000000000000000000000000"
	UncommittedSID = "000000"
)

type Repo struct {
	RootPath string
	Commits  []Commit
	Branches []Branch
	Status   Status
}

type Git interface {
	GetRepo() (Repo, error)
	RepoPath() string
	GetLog() ([]Commit, error)
	GetStatus() (Status, error)
	GetBranches() ([]Branch, error)

	IsIgnored(path string) bool
	CommitDiff(id string) (CommitDiff, error)
	Checkout(name string) error
	Commit(message string) error
	Fetch() error
	PushBranch(name string) error
	CreateBranch(name string) error
	MergeBranch(name string) error
	DeleteRemoteBranch(name string) error
	DeleteLocalBranch(name string) error
}

type git struct {
	cmd           gitCommander
	statusService *statusService
	logService    *logService
	branchService *branchesService
	ignoreService *ignoreService
	diffService   *diffService
	commitService *commitService
	remoteService *remoteService
}

func NewGit(path string) Git {
	cmd := newGitCmd(path)
	status := newStatus(cmd)
	return &git{
		cmd:           cmd,
		statusService: status,
		logService:    newLog(cmd),
		branchService: newBranchService(cmd),
		remoteService: newRemoteService(cmd),
		ignoreService: newIgnoreHandler(path),
		diffService:   newDiff(cmd, status),
		commitService: newCommit(cmd),
	}
}
func (h *git) GetRepo() (Repo, error) {
	commits, err := h.logService.getLog()
	if err != nil {
		return Repo{}, err
	}
	branches, err := h.branchService.getBranches()
	if err != nil {
		return Repo{}, err
	}
	status, err := h.statusService.getStatus()
	if err != nil {
		return Repo{}, err
	}

	return Repo{RootPath: h.cmd.RepoPath(), Commits: commits, Branches: branches, Status: status}, nil
}

func (h *git) RepoPath() string {
	return h.cmd.RepoPath()
}

func (h *git) GetLog() ([]Commit, error) {
	return h.logService.getLog()
}

func (h *git) GetBranches() ([]Branch, error) {
	return h.branchService.getBranches()
}

func (h *git) GetStatus() (Status, error) {
	return h.statusService.getStatus()
}

func (h *git) Fetch() error {
	return h.remoteService.fetch()
}

func (h *git) CommitDiff(id string) (CommitDiff, error) {
	return h.diffService.commitDiff(id)
}

func (h *git) IsIgnored(path string) bool {
	return h.ignoreService.isIgnored(path)
}

func (h *git) Checkout(name string) error {
	return h.branchService.checkout(name)
}

func (h *git) Commit(message string) error {
	return h.commitService.commitAllChanges(message)
}

func (h *git) PushBranch(name string) error {
	return h.remoteService.pushBranch(name)
}

func (h *git) MergeBranch(name string) error {
	return h.branchService.mergeBranch(name)
}

func (h *git) CreateBranch(name string) error {
	return h.branchService.createBranch(name)
}

func (h *git) DeleteRemoteBranch(name string) error {
	return h.remoteService.deleteRemoteBranch(name)
}
func (h *git) DeleteLocalBranch(name string) error {
	return h.branchService.deleteLocalBranch(name)
}

// GitVersion returns the git version
func Version() string {
	out, _ := exec.Command("git", "version").Output()
	return strings.TrimSpace(string(out))
}
