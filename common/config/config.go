package config

import (
	"encoding/json"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
	"os"
	"path/filepath"
	"sync"
)

type Config struct {
	ReleasesEtag  string
	StableRelease Release
	PreRelease    Release
	Repos         []Repo
}

type Repo struct {
	Path     string
	Branches []Branch
}

type Branch struct {
	DisplayName string
	Color       int
}

//
type Release struct {
	Version string
	Assets  []Asset
}

type Asset struct {
	Name string
	Url  string
}

type Service struct {
	lock          sync.Mutex
	currentConfig Config
}

func NewConfig() *Service {
	return &Service{}
}

func (s *Service) Load() {
	s.currentConfig = s.defaultConfig()
	configFile, err := utils.FileRead(s.configPath())
	if err != nil {
		log.Infof("No config at %s, %v", s.configPath(), err)
		s.save(s.currentConfig)
		return
	}

	var config Config
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		log.Warnf("Failed to parse config at %s, %v", s.configPath(), err)
		return
	}
	s.currentConfig = config
}

func (s *Service) Get() Config {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.currentConfig
}

func (s *Service) Set(setFunc func(c *Config)) {
	config := s.Get()
	setFunc(&config)
	s.save(config)
}

func (s *Service) GetRepo(path string) Repo {
	config := s.Get()
	log.Infof("Config %#v", config)
	log.Infof("get repo for %q", path)
	for _, repo := range config.Repos {
		if path == repo.Path {
			return repo
		}
	}
	return s.defaultRepo(path)
}

func (s *Service) SetRepo(path string, setFunc func(r *Repo)) {
	config := s.Get()
	repo := s.defaultRepo(path)
	index := -1
	for i, r := range config.Repos {
		if path == r.Path {
			repo = r
			index = i
			break
		}
	}
	if index == -1 {
		// New repo, add to config
		index = len(config.Repos)
		config.Repos = append(config.Repos, repo)
	}

	setFunc(&repo)
	config.Repos[index] = repo
	s.save(config)
}

func (s *Service) configPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(log.Error(err))
	}
	return filepath.Join(home, ".gmc")
}

func (s *Service) save(config Config) {
	configFile := utils.MustJsonMarshal(config)
	utils.MustFileWrite(s.configPath(), configFile)
	s.lock.Lock()
	defer s.lock.Unlock()
	s.currentConfig = config
}

func (s *Service) defaultConfig() Config {
	return Config{}
}

func (s *Service) defaultRepo(path string) Repo {
	return Repo{Path: path}
}
