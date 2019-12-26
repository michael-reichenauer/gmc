package gitlib

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils/log"
	"os/exec"
)

type gitCmd struct {
	workingDir string
}

func newGitCmd(workingDir string) *gitCmd {
	return &gitCmd{workingDir: workingDir}
}

func (h *gitCmd) git(arg ...string) (string, error) {
	cmd := exec.Command("git", arg...)
	cmd.Dir = h.workingDir

	// Get the git cmd output
	out, err := cmd.Output()
	if err != nil {
		msg := fmt.Sprintf("failed to do git %v in %v, %v", arg, h.workingDir, err)
		log.Warnf(msg)
		return "", fmt.Errorf(msg)
	}
	return string(out), nil
}
