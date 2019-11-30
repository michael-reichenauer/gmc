package git

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils/log"
	"os/exec"
)

func gitCmd(workingDir string, arg ...string) (string, error) {
	cmd := exec.Command("git", arg...)
	cmd.Dir = workingDir

	// Get the git cmd output
	out, err := cmd.Output()
	if err != nil {
		msg := fmt.Sprintf("failed to do git %v in %v, %v", arg, workingDir, err)
		log.Warnf(msg)
		return "", fmt.Errorf(msg)
	}
	return string(out), nil
}
