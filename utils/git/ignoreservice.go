package git

import (
	"github.com/bmatcuk/doublestar"
	"github.com/michael-reichenauer/gmc/utils"
	"path/filepath"
	"strings"
)

type ignoreService struct {
	dirPrefixLen int
	patters      []string
}

func newIgnoreHandler(rootPath string) *ignoreService {
	h := &ignoreService{dirPrefixLen: len(rootPath)}
	h.parseIgnoreFile(rootPath)
	return h
}

func (t *ignoreService) isIgnored(path string) bool {
	if filepath.IsAbs(path) && len(path) > t.dirPrefixLen {
		path = path[t.dirPrefixLen+1:]
	}
	for _, pattern := range t.patters {
		match, err := doublestar.Match(pattern, path)
		if err != nil {
			return false
		}
		if match {
			return true
		}
	}
	return false
}

func (t *ignoreService) parseIgnoreFile(repoPath string) {
	ignorePath := filepath.Join(repoPath, ".gitignore")
	file, err := utils.FileRead(ignorePath)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(file), "\n") {
		if index := strings.Index(line, "#"); index != -1 {
			line = line[:index]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasSuffix(line, "/") {
			line = line + "**"
		}
		t.patters = append(t.patters, line)
	}
}
