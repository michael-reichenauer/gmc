package main

import (
	"flag"
	"fmt"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/installation"
	"github.com/michael-reichenauer/gmc/program"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/log/logger"
	"github.com/michael-reichenauer/gmc/utils/timer"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"io/ioutil"
	stdlog "log"
	_ "net/http/pprof"
	"os"
	"os/exec"
	"runtime"
)

const (
	version = "0.31"
)

var (
	workingFolderFlag = flag.String("d", "", "specify working folder")
	showVersionFlag   = flag.Bool("version", false, "print gmc version")
	pauseFlag         = flag.Bool("pause", false, "pause until user click enter")
	externalWindow    = flag.Bool("external", false, "start gmc in external window (used by ide)")
)

func main() {
	st := timer.Start()
	flag.Parse()
	if *showVersionFlag {
		fmt.Printf("%s", version)
		return
	}

	if isWindowsGolandConsole() {
		// Seems program is run within a debugger console, which does not support termbox
		// So a new external process is started and this instance ends
		startAsExternalProcess()
		return
	}

	// go func() {
	// 	log.Infof("prof on port 6060 %v", http.ListenAndServe("localhost:6060", nil))
	// }()

	if *pauseFlag {
		// The process was started with 'pause' flag, e.g. from Goland,
		// wait until enter is clicked, this can be used for attaching a debugger
		fmt.Printf("Attach a debugger and click 'enter' to proceed ...\n")
		utils.ReadLine()
	}

	// Disable standard logging since some modules log to stderr, which conflicts with console ui
	stdlog.SetOutput(ioutil.Discard)
	// Set default http client proxy to the system proxy (used by e.g. telemetry)
	utils.SetDefaultHTTPProxy()
	logger.StdTelemetry.Enable(version)
	defer logger.StdTelemetry.Close()
	log.Eventf("program-start", "Starting gmc %s ...", version)

	logger.RedirectStdErrorToFile()
	defer log.Event("program-stop")
	configService := config.NewConfig(version, *workingFolderFlag, "")
	configService.SetState(func(s *config.State) {
		s.InstalledVersion = version
	})

	logProgramInfo(configService)
	autoUpdate := installation.NewAutoUpdate(configService, version)
	autoUpdate.Start()

	uiHandler := ui.NewUI()
	uiHandler.Run(func() {
		log.Infof("Show main window %s", st)
		mainWindow := program.NewMainWindow(uiHandler, configService)
		mainWindow.Show()
	})
}

func isWindowsGolandConsole() bool {
	return runtime.GOOS == "windows" && *externalWindow
	// // termbox requires a "real" terminal. The GoLand console is not supported.
	// // so check if current terminal is ok (check only on windows)
	// if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) ||
	// 	runtime.GOOS != "windows" {
	// 	return false
	// }
	// return true
}

func startAsExternalProcess() {
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

func logProgramInfo(configService *config.Service) {
	log.Infof("Version: %s", configService.ProgramVersion)
	log.Infof("Build: release=%v", program.IsRelease)
	log.Infof("Binary path: %q", utils.BinPath())
	log.Infof("Args: %v", os.Args)
	log.Infof("OS: %q", runtime.GOOS)
	log.Infof("Arch: %q", runtime.GOARCH)
	log.Infof("Go version: %q", runtime.Version())
	log.Infof("Folder path: %q", configService.FolderPath)
	log.Infof("Working Folder: %q", utils.CurrentDir())
	log.Infof("Http proxy: %q", utils.GetHTTPProxyURL())
	go func() { log.Infof("Git version: %q", git.Version()) }()
}
