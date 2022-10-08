package git

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"

	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/timer"
)

type gitCommander interface {
	Git(arg ...string) (string, error)
	WorkingDir() string
	ReadFile(path string) (string, error)
}

type gitCmd struct {
	workingDir string
}

func newGitCmd(workingDir string) gitCommander {
	return &gitCmd{workingDir: workingDir}
}

func (t *gitCmd) WorkingDir() string {
	return t.workingDir
}

func (t *gitCmd) ReadFile(path string) (string, error) {
	bytes, err := ioutil.ReadFile(path)
	return string(bytes), err
}

func (t *gitCmd) Git(args ...string) (string, error) {
	argsText := strings.Join(args, " ")
	log.Debugf("Cmd: git %s (%s) ...", argsText, t.workingDir)
	// Get the git cmd output
	st := timer.Start()
	c := exec.Command("git", args...)
	c.Dir = t.workingDir
	out, err := c.Output()
	if err != nil {
		errorText := ""
		if ee, ok := err.(*exec.ExitError); ok {
			errorText = string(ee.Stderr)
			errorText = strings.ReplaceAll(errorText, "\t", "   ")
		}
		err := fmt.Errorf("failed: git %s\n%v\n%v", argsText, err, errorText)
		log.Warnf("%v %v", err, st)
		return string(out), err
	}
	log.Infof("OK: git %s %v", argsText, st)
	return string(out), nil
}
