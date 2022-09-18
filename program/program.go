package program

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
)

func LogProgramInfo(version, workingDirFlag string) {
	log.Infof("Version: %s", version)
	log.Infof("Build: release=%v", IsRelease)
	log.Infof("Binary path: %q", utils.BinPath())
	log.Infof("Args: %v", os.Args)
	log.Infof("ID: %q", utils.MachineID)
	log.Infof("OS: %q", runtime.GOOS)
	log.Infof("Arch: %q", runtime.GOARCH)
	log.Infof("Go version: %q", runtime.Version())
	log.Infof("Specified folder path: %q", workingDirFlag)
	log.Infof("Working Folder: %q", utils.CurrentDir())
	log.Infof("Http proxy: %q", utils.GetHTTPProxyURL())
	go func() { log.Infof("Git version: %q", git.Version()) }()
}

func IsWindowsGolandConsole(externalWindow, pauseFlag bool) bool {
	return runtime.GOOS == "windows" && externalWindow && !pauseFlag
	// // termbox requires a "real" terminal. The GoLand console is not supported.
	// // so check if current terminal is ok (check only on windows)
	// if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) ||
	// 	runtime.GOOS != "windows" {
	// 	return false
	// }
	// return true
}

func StartAsExternalProcess() {
	args := []string{"/C", "start"}
	args = append(args, os.Args...)
	i := utils.StringsIndex(args, "-external")
	if i != -1 {
		// Remove the external flag before restarting
		args = append(args[:i], args[i+1:]...)
	}
	cmd := exec.Command("cmd", args...)
	_ = cmd.Start()
	_ = cmd.Wait()
}
