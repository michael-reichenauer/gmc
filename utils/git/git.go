package git

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/michael-reichenauer/gmc/utils"
)

const (
	UncommittedID      = "0000000000000000000000000000000000000000"
	UncommittedSID     = "000000"
	PartialLogCommitID = "ffffffffffffffffffffffffffffffffffffffff"
)

var ErrConflicts = errors.New("merge resulted in conflict(s)")

type Repo struct {
	RootPath string
	Commits  []Commit
	Branches []Branch
	Status   Status
	Tags     []Tag
}

type Git interface {
	GetRepo(maxCommitCount int) (Repo, error)
	RepoPath() string
	GetLog() (Commits, error)
	GetStatus() (Status, error)
	GetBranches() (Branches, error)
	GetFiles(ref string) ([]string, error)

	InitRepo() error
	InitRepoBare() error
	Clone(uri, path string) error
	ConfigUser(name, email string) error

	IsIgnored(path string) bool
	CommitDiff(id string) (CommitDiff, error)
	FileDiff(path string) ([]CommitDiff, error)
	Checkout(name string) error
	Commit(message string) error
	Fetch() error
	PushBranch(name string) error
	CreateBranch(name string) error
	CreateBranchAt(name string, id string) error
	MergeBranch(name string) error
	DeleteRemoteBranch(name string) error
	DeleteLocalBranch(name string) error
	GetTags() ([]Tag, error)
	PullCurrentBranch() error
	PullBranch(name string) error
	GetValue(key string) (string, error)
	SetValue(key, value string) error
}

type git struct {
	cmd             gitCommander
	statusService   *statusService
	logService      *logService
	branchService   *branchesService
	ignoreService   *ignoreService
	diffService     *diffService
	commitService   *commitService
	remoteService   *remoteService
	tagService      *tagService
	keyValueService *keyValueService
	repoService     *repoService
	configService   *configService
}

func New(path string) Git {
	cmd := newGitCmd(path)
	return NewWithCmd(cmd)
}

func NewWithCmd(cmd gitCommander) Git {
	status := newStatus(cmd)
	return &git{
		cmd:             cmd,
		statusService:   status,
		logService:      newLog(cmd),
		branchService:   newBranchService(cmd),
		remoteService:   newRemoteService(cmd),
		ignoreService:   newIgnoreHandler(cmd.WorkingDir()),
		diffService:     newDiff(cmd, status),
		commitService:   newCommit(cmd),
		tagService:      newTagService(cmd),
		keyValueService: newKeyValue(cmd),
		repoService:     newRepoService(cmd),
		configService:   newConfigService(cmd),
	}
}

func (t *git) InitRepo() error {
	return t.repoService.InitRepo()
}

func (t *git) InitRepoBare() error {
	return t.repoService.InitRepoBare()
}

func (t *git) Clone(uri, path string) error {
	return t.remoteService.clone(uri, path)
}

func (t *git) ConfigUser(name, email string) error {
	return t.configService.ConfigUser(name, email)
}

func (t *git) GetRepo(maxCommitCount int) (Repo, error) {
	commits, err := t.logService.getLog(maxCommitCount)
	if err != nil {
		return Repo{}, err
	}
	branches, err := t.branchService.getBranches()
	if err != nil {
		return Repo{}, err
	}
	status, err := t.statusService.getStatus()
	if err != nil {
		return Repo{}, err
	}
	tags, err := t.tagService.getTags()
	if err != nil {
		return Repo{}, err
	}

	return Repo{
		RootPath: t.cmd.WorkingDir(),
		Commits:  commits,
		Branches: branches,
		Status:   status,
		Tags:     tags,
	}, nil
}

func (t *git) RepoPath() string {
	return t.cmd.WorkingDir()
}

func (t *git) GetLog() (Commits, error) {
	return t.logService.getLog(-1)
}

func (t *git) GetBranches() (Branches, error) {
	return t.branchService.getBranches()
}

func (t *git) GetFiles(ref string) ([]string, error) {
	return t.logService.getFiles(ref)
}

func (t *git) GetStatus() (Status, error) {
	return t.statusService.getStatus()
}

func (t *git) Fetch() error {
	return t.remoteService.fetch()
}

func (t *git) CommitDiff(id string) (CommitDiff, error) {
	return t.diffService.commitDiff(id)
}

func (t *git) FileDiff(path string) ([]CommitDiff, error) {
	return t.diffService.fileDiff(path)
}

func (t *git) IsIgnored(path string) bool {
	return t.ignoreService.isIgnored(path)
}

func (t *git) Checkout(name string) error {
	return t.branchService.checkout(name)
}

func (t *git) Commit(message string) error {
	return t.commitService.commitAllChanges(message)
}

func (t *git) PushBranch(name string) error {
	return t.remoteService.pushBranch(name)
}

func (t *git) PullCurrentBranch() error {
	return t.remoteService.pullCurrentBranch()
}

func (t *git) PullBranch(name string) error {
	return t.remoteService.pullBranch(name)
}

func (t *git) MergeBranch(name string) error {
	return t.branchService.mergeBranch(name)
}

func (t *git) CreateBranch(name string) error {
	return t.branchService.createBranch(name)
}

func (t *git) CreateBranchAt(name string, id string) error {
	return t.branchService.createBranchAt(name, id)
}

func (t *git) DeleteRemoteBranch(name string) error {
	return t.remoteService.deleteRemoteBranch(name)
}

func (t *git) DeleteLocalBranch(name string) error {
	return t.branchService.deleteLocalBranch(name)
}

func (t *git) GetTags() ([]Tag, error) {
	return t.tagService.getTags()
}

// GitVersion returns the git version
func Version() string {
	out, _ := exec.Command("git", "version").Output()
	return strings.TrimSpace(string(out))
}

func (t *git) GetValue(key string) (string, error) {
	return t.keyValueService.getValue(key)
}

func (t *git) SetValue(key, value string) error {
	return t.keyValueService.setValue(key, value)
}

func StripRemotePrefix(name string) string {
	return strings.TrimPrefix(name, "origin/")
}

func WorkingTreeRoot(path string) (string, error) {
	current := path
	if strings.HasSuffix(path, ".git") || strings.HasSuffix(path, ".git/") || strings.HasSuffix(path, ".git\\") {
		current = filepath.Dir(path)
	}

	for {
		gitRepoPath := filepath.Join(current, ".git")
		if utils.DirExists(gitRepoPath) {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			// Reached top/root volume folder
			break
		}
		current = parent
	}
	return "", fmt.Errorf("could not locate git repo in or above " + path)
}
