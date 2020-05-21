package server

import (
	"github.com/michael-reichenauer/gmc/api"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/server/viewrepo"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/git"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
)

type server struct {
	configService *config.Service
}

func NewServer(configService *config.Service) api.Api {
	return &server{configService: configService}
}

func (t *server) GetSubDirs(parentDirPath string) ([]string, error) {
	var paths []string
	if parentDirPath == "" {
		// Path not specified, return recent used parent paths and root folders
		paths = t.configService.GetState().RecentParentFolders
		paths = append(paths, utils.GetVolumes()...)
		return paths, nil
	}

	return t.getSubDirs(parentDirPath), nil
}

func (t *server) OpenRepo(path string) (api.Repo, error) {
	if path == "" {
		// No path specified, assume current working dir
		path = utils.CurrentDir()
	}
	workingDir, err := git.WorkingDirRoot(path)
	if err != nil {
		// Could not locate a working dir root
		return nil, err
	}

	parentDir := filepath.Dir(workingDir)
	t.configService.SetState(func(s *config.State) {
		s.RecentFolders = utils.RecentItems(s.RecentFolders, workingDir, 10)
		s.RecentParentFolders = utils.RecentItems(s.RecentParentFolders, parentDir, 5)
	})

	viewRepo := viewrepo.NewViewRepo(t.configService, workingDir)
	return viewRepo, nil
}

func (t *server) GetRecentDirs() ([]string, error) {
	return t.configService.GetState().RecentFolders, nil
}

func (t *server) getSubDirs(path string) []string {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		// Folder not readable, might be e.g. access denied
		return nil
	}

	var paths []string
	for _, f := range files {
		if !f.IsDir() || f.Name() == "$RECYCLE.BIN" {
			continue
		}
		paths = append(paths, filepath.Join(path, f.Name()))
	}
	// Sort with but ignore case
	sort.SliceStable(paths, func(l, r int) bool {
		return -1 == strings.Compare(strings.ToLower(paths[l]), strings.ToLower(paths[r]))
	})
	return paths
}
