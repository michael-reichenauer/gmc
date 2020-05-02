package git

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/timer"
	"io/ioutil"
	"os/exec"
	"strings"
)

type gitCommander interface {
	Git(arg ...string) (string, error)
	RepoPath() string
	ReadFile(path string) (string, error)
}

type gitCmd struct {
	repoPath string
}

func newGitCmd(repoPath string) gitCommander {
	rootPath, err := WorkingFolderRoot(repoPath)
	if err == nil {
		repoPath = rootPath
	}
	return &gitCmd{repoPath: repoPath}
}

func newGitCmdWithRoot(repoPath string) gitCommander {
	return &gitCmd{repoPath: repoPath}
}

func (t *gitCmd) RepoPath() string {
	return t.repoPath
}

func (t *gitCmd) ReadFile(path string) (string, error) {
	bytes, err := ioutil.ReadFile(path)
	return string(bytes), err
}

func (t *gitCmd) Git(args ...string) (string, error) {
	argsText := strings.Join(args, " ")
	log.Infof("Cmd: git %s (%s) ...", argsText, t.repoPath)
	// Get the git cmd output
	st := timer.Start()
	c := exec.Command("git", args...)
	c.Dir = t.repoPath
	out, err := c.Output()
	if err != nil {
		errorText := ""
		if ee, ok := err.(*exec.ExitError); ok {
			errorText = string(ee.Stderr)
			errorText = strings.ReplaceAll(errorText, "\t", "   ")
		}
		err := fmt.Errorf("error: git %s\n%v\n%v", argsText, err, errorText)
		log.Warnf("%v %v", err, st)
		return string(out), err
	}
	log.Infof("OK: git %s %v", argsText, st)
	return string(out), nil
}
