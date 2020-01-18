package program

import (
	"flag"
	"fmt"
	"github.com/mattn/go-isatty"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/installation"
	"github.com/michael-reichenauer/gmc/repoview"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/gitlib"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/log/logger"
	"github.com/michael-reichenauer/gmc/utils/ui"
	"io/ioutil"
	stdlog "log"
	"os"
	"os/exec"
	"runtime"
)

var (
	repoPathFlag = flag.String("d", "", "specify working folder")
	versionFlag  = flag.Bool("version", false, "print gmc version")
)

func Main(version string) {
	flag.Parse()
	if *versionFlag {
		fmt.Printf("%s", version)
		return
	}

	if isDebugConsole() {
		// Seems program is run within a debugger console, which the termbox does not support
		// So a new external process has been started and this instance ends
		return
	}
	// Disable standard logging since some modules log to stderr, which conflicts with console ui
	stdlog.SetOutput(ioutil.Discard)
	// Set default http client proxy to the system proxy (used by e.g. telemetry)
	utils.SetDefaultHTTPProxy()
	logger.StdTelemetry.Enable(version)

	log.Eventf("program-start", "Starting gmc %s ...", version)

	configService := config.NewConfig()
	configService.SetState(func(s *config.State) {
		s.InstalledVersion = version
	})
	workingFolderPath := getWorkingFolder()

	logProgramInfo(version, workingFolderPath)

	autoUpdate := installation.NewAutoUpdate(configService, version)
	autoUpdate.Start()
	//autoUpdate.UpdateIfAvailable()

	uiHandler := ui.NewUI()
	uiHandler.Run(func() {
		mainWindow := repoview.NewMainWindow(uiHandler, configService, workingFolderPath, version)
		uiHandler.OnResizeWindow = mainWindow.OnResizeWindow
		mainWindow.Show()
	})

	log.Event("program-stop")
	logger.StdTelemetry.Close()
	log.Infof("Exit gmc")
}

func isDebugConsole() bool {
	if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) ||
		runtime.GOOS != "windows" {
		return false
	}

	// Seems to be not running in a terminal like e.g. in goland,
	// termbox requires a terminal, so lets restart as an external command on windows
	args := []string{"/C", "start"}
	args = append(args, os.Args...)
	cmd := exec.Command("cmd", args...)
	_ = cmd.Start()
	_ = cmd.Wait()
	return true
}

func getWorkingFolder() string {
	if *repoPathFlag == "" {
		// No specified repo path, use current dir
		*repoPathFlag = utils.CurrentDir()
	}
	path, err := gitlib.WorkingFolderRoot(*repoPathFlag)
	if err != nil {
		panic(log.Fatal(err))
	}
	return path
}

func logProgramInfo(version string, path string) {
	log.Infof("Version: %s", version)
	log.Infof("Binary path: %q", utils.BinPath())
	log.Infof("Http proxy: %s", utils.GetHTTPProxyURL())
	log.Infof("Working folder: %s", path)
}
