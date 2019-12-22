package git

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils"
	"path/filepath"
	"strings"
)

type Repo struct {
	status     *statusHandler
	logHandler *logHandler
	branches   *branchesHandler
	RepoPath   string
}

func NewRepo(path string) *Repo {
	rootPath, err := WorkingFolderRoot(path)
	if err == nil {
		path = rootPath
	}
	cmd := newGitCmd(path)
	return &Repo{
		status:     newStatus(cmd),
		logHandler: newLog(cmd),
		branches:   newBranches(cmd),
		RepoPath:   path,
	}
}

func (h *Repo) GetLog() ([]Commit, error) {
	return h.logHandler.getLog()
}

func (h *Repo) GetBranches() ([]Branch, error) {
	return h.branches.getBranches()
}

func (h *Repo) GetStatus() (Status, error) {
	return h.status.getStatus()
}

func WorkingFolderRoot(path string) (string, error) {
	current := path
	if strings.HasSuffix(path, ".git") || strings.HasSuffix(path, ".git/") || strings.HasSuffix(path, ".git\\") {
		current = filepath.Dir(path)
	}

	for current != "" {
		gitRepoPath := filepath.Join(current, ".git")
		if utils.DirExists(gitRepoPath) {
			return current, nil
		}
		current = filepath.Dir(current)
	}
	return "", fmt.Errorf("could not locater working folder root from " + path)
}
