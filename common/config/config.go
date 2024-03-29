package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
)

const (
	stateName  = ".gmcstate"
	configName = ".gmcconfig"
)

type Config struct {
	DisableAutoUpdate bool
	AllowPreview      bool
}

type State struct {
	InstalledVersion    string
	ReleasesEtag        string
	StableRelease       Release
	PreRelease          Release
	Repos               []Repo
	RecentFolders       []string
	RecentParentFolders []string
}

type Repo struct {
	Path             string
	ShownBranches    []string
	BranchesChildren map[string][]string
	// Branches      []Branch
}

type Branch struct {
	DisplayName string
	Color       int
}

type Release struct {
	Version string
	Preview bool
	Assets  []Asset
}

type Asset struct {
	Name string
	Url  string
}

type Service struct {
	ProgramVersion string
	dataFolder     string
}

func NewConfig(programVersion, dataFolder string) *Service {
	return &Service{ProgramVersion: programVersion, dataFolder: dataFolder}
}

func (s *Service) GetState() State {
	current := s.defaultState()
	stateFile, err := utils.FileRead(s.statePath())
	if err != nil {
		log.Infof("No state at %s, %v", s.statePath(), err)
		s.saveState(current)
		return current
	}

	var state State
	err = json.Unmarshal(stateFile, &state)
	if err != nil {
		log.Warnf("Failed to parse %s, %v", s.statePath(), err)
		return current
	}
	return state
}

func (s *Service) SetState(setFunc func(s *State)) {
	state := s.GetState()
	setFunc(&state)
	s.saveState(state)
}

func (s *Service) SetConfig(setFunc func(c *Config)) {
	current := s.GetConfig()
	setFunc(&current)
	s.saveConfig(current)
}

func (s *Service) GetConfig() Config {
	current := s.defaultConfig()
	configFile, err := utils.FileRead(s.configPath())
	if err != nil {
		log.Infof("No config at %s, %v", s.statePath(), err)
		s.saveConfig(current)
		return current
	}

	var config Config
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		log.Warnf("Failed to parse %s, %v", s.configPath(), err)
		return current
	}
	return config
}

func (s *Service) GetRepo(path string) Repo {
	config := s.GetState()
	log.Debugf("Config %#v", config)
	for _, repo := range config.Repos {
		if path == repo.Path {
			if repo.BranchesChildren == nil {
				repo.BranchesChildren = make(map[string][]string)
			}
			return repo
		}
	}
	return s.defaultRepo(path)
}

func (s *Service) SetRepo(path string, setFunc func(r *Repo)) {
	state := s.GetState()
	repo := s.defaultRepo(path)
	index := -1
	for i, r := range state.Repos {
		if path == r.Path {
			repo = r
			index = i
			break
		}
	}
	if index == -1 {
		// New repo, add to config
		index = len(state.Repos)
		state.Repos = append(state.Repos, repo)
	}

	if repo.BranchesChildren == nil {
		repo.BranchesChildren = make(map[string][]string)
	}

	setFunc(&repo)
	state.Repos[index] = repo
	s.saveState(state)
}

func (s *Service) statePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(log.Fatal(err))
	}
	return filepath.Join(home, stateName)
}

func (s *Service) saveState(state State) {
	file, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		panic(log.Fatal(err))
	}
	utils.MustFileWrite(s.statePath(), file)
}

func (s *Service) defaultState() State {
	return State{}
}

func (s *Service) configPath() string {
	dataFolder := s.dataFolder
	if dataFolder == "" {
		dataFolder = utils.HomeDir()
	}
	return filepath.Join(dataFolder, configName)
}

func (s *Service) saveConfig(config Config) {
	file, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		panic(log.Fatal(err))
	}
	utils.MustFileWrite(s.configPath(), file)
}

func (s *Service) defaultConfig() Config {
	return Config{AllowPreview: true}
}

func (s *Service) defaultRepo(path string) Repo {
	return Repo{Path: path, BranchesChildren: make(map[string][]string)}
}
