package git

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	lock      sync.Mutex
	traceDir  string
	replayDir string
)

type GitCommander interface {
	Git(arg ...string) (string, error)
	RepoPath() string
	ReadFile(path string) (string, error)
}

type gitCmd struct {
	repoPath string
}

type command struct {
	RepoPath string
	Name     string
	Args     []string
	Err      string
	Output   string
}

func newGitCmd(repoPath string) *gitCmd {
	rootPath, err := WorkingFolderRoot(repoPath)
	if err == nil {
		repoPath = rootPath
	}
	return &gitCmd{repoPath: repoPath}
}

func EnableTracing(name string) {
	lock.Lock()
	defer lock.Unlock()
	path := TracePath(name)
	_ = os.RemoveAll(path)
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		panic(log.Fatal(err))
	}
	traceDir = path
}

func DisableTracing(name string) {
	lock.Lock()
	defer lock.Unlock()
	traceDir = ""
}

func EnableReplay(name string) {
	lock.Lock()
	defer lock.Unlock()
	path := TracePath(name)
	replayDir = path
}

func DisableReplay(name string) {
	lock.Lock()
	defer lock.Unlock()
	replayDir = ""
}

func CurrentTracePath() string {
	lock.Lock()
	defer lock.Unlock()
	return traceDir
}

func TracePath(name string) string {
	var path string
	if name == "" {
		path = filepath.Join(os.TempDir(), "gmctrace")
	} else {
		path = filepath.Join(os.TempDir(), name)
	}
	return path
}

func (h *gitCmd) RepoPath() string {
	return h.repoPath
}

func (h *gitCmd) ReadFile(path string) (string, error) {
	cmd := command{RepoPath: h.repoPath, Name: "ReadFile", Args: []string{path}}
	return h.runCommand(cmd)
}

func (h *gitCmd) Git(arg ...string) (string, error) {
	cmd := command{RepoPath: h.repoPath, Name: "git", Args: arg}
	return h.runCommand(cmd)
}

func (h *gitCmd) runCommand(cmd command) (string, error) {
	lock.Lock()
	if traceDir == "" && replayDir != "" {
		fileName := h.fileName(cmd)
		path := filepath.Join(replayDir, fileName)
		cmdBytes, err := ioutil.ReadFile(path)
		if err != nil {
			panic(log.Fatal(err))
		}
		err = json.Unmarshal(cmdBytes, &cmd)
		if err != nil {
			panic(log.Fatal(err))
		}
		log.Infof("Read %d bytes to %s", len(cmdBytes), path)
		lock.Unlock()
		if cmd.Err != "" {
			return "", fmt.Errorf(cmd.Err)
		}
		return cmd.Output, nil
	}
	lock.Unlock()

	switch cmd.Name {
	case "git":
		cmd = h.runGitCommand(cmd)
	case "ReadFile":
		cmd = h.runReadFileCommand(cmd)
	default:
		panic(log.Fatal(fmt.Errorf("unknown command %s", cmd.Name)))
	}

	lock.Lock()
	if replayDir == "" && traceDir != "" {
		fileName := h.fileName(cmd)
		cmdBytes, err := json.Marshal(cmd)
		if err != nil {
			panic(log.Fatal(err))
		}
		path := filepath.Join(traceDir, fileName)
		err = ioutil.WriteFile(path, cmdBytes, 0644)
		if err != nil {
			panic(log.Fatal(err))
		}
		log.Infof("Wrote %d bytes to %s", len(cmdBytes), path)
	}
	lock.Unlock()

	if cmd.Err != "" {
		return "", fmt.Errorf(cmd.Err)
	}
	return cmd.Output, nil
}

func (h *gitCmd) fileName(cmd command) string {
	return fmt.Sprintf("%s_%x", cmd.Name, sha256.Sum256([]byte(fmt.Sprintf("%s %v", cmd.Name, cmd.Args))))
}
func (h *gitCmd) runGitCommand(cmd command) command {
	log.Infof("Cmd: %s %s (%s) ...", cmd.Name, strings.Join(cmd.Args, " "), cmd.RepoPath)
	//fmt.Printf("Cmd: %s %s (%s) ...\n", cmd.Name, strings.Join(cmd.Args, " "), cmd.RepoPath)
	// Get the git cmd output
	t := time.Now()
	c := exec.Command(cmd.Name, cmd.Args...)
	c.Dir = h.repoPath
	out, err := c.Output()
	if err != nil {
		errorText := ""
		if ee, ok := err.(*exec.ExitError); ok {
			errorText = string(ee.Stderr)
			errorText = strings.ReplaceAll(errorText, "\t", "   ")
		}
		msg := fmt.Sprintf("error: git %s\n%v\n%v", strings.Join(cmd.Args, " "), err, errorText)
		log.Warnf("%s (%v)", msg, time.Since(t))
		cmd.Err = msg
		return cmd
	}
	cmd.Output = string(out)
	log.Infof("OK: git %s (%v)", strings.Join(cmd.Args, " "), time.Since(t))
	return cmd
}

func (h *gitCmd) runReadFileCommand(cmd command) command {
	bytes, err := ioutil.ReadFile(cmd.Args[0])
	cmd.Output = string(bytes)
	if err != nil {
		cmd.Err = err.Error()
	}
	return cmd
}

func WorkingFolderRoot(path string) (string, error) {
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
	return "", fmt.Errorf("could not locater working folder root from " + path)
}
