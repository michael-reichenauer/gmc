package gitmodel

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils"
	"path/filepath"
	"strings"
)

func GetWorkingFolderRoot(path string) (string, error) {
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
