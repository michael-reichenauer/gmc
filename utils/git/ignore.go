package git

import (
	"github.com/bmatcuk/doublestar"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
	"path/filepath"
	"strings"
)

//import "github.com/bmatcuk/doublestar"

type ignoreHandler struct {
	dirPrefixLen int
	patters      []string
}

func newIgnoreHandler(repoPath string) *ignoreHandler {
	rootPath, err := WorkingFolderRoot(repoPath)
	if err == nil {
		repoPath = rootPath
	}

	h := &ignoreHandler{dirPrefixLen: len(repoPath)}
	h.parseIgnoreFile(repoPath)
	return h
}

func (h *ignoreHandler) isIgnored(path string) bool {
	if filepath.IsAbs(path) && len(path) > h.dirPrefixLen {
		path = path[h.dirPrefixLen+1:]
	}
	if strings.Contains(path, "shelf") {
		log.Infof("")
	}
	for _, pattern := range h.patters {
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

func (h *ignoreHandler) parseIgnoreFile(repoPath string) {
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
		h.patters = append(h.patters, line)
	}
}
