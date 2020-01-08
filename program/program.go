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
	"github.com/michael-reichenauer/gmc/utils/telemetry"
	"github.com/michael-reichenauer/gmc/utils/ui"
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
		return
	}

	proxyURL := utils.SetDefaultHTTPProxy()
	tel := telemetry.NewTelemetry(version)
	logger.Std.SetTelemetry(tel)

	log.Infof("Starting gmc version: %s %q ...", version, utils.BinPath())
	tel.SendEventf("program-start", "Starting gmc version: %s %q ...", version, utils.BinPath())
	log.Infof("Using default http proxy:%q", proxyURL)

	configService := config.NewConfig()
	configService.SetState(func(s *config.State) {
		s.InstalledVersion = version
	})
	autoUpdate := installation.NewAutoUpdate(configService, tel, version)
	autoUpdate.Start()
	//autoUpdate.UpdateIfAvailable()

	if *repoPathFlag == "" {
		// No specified repo path, use current dir
		*repoPathFlag = utils.CurrentDir()
	}

	path, err := gitlib.WorkingFolderRoot(*repoPathFlag)
	if err != nil {
		panic(log.Fatal(err))
	}
	log.Infof("Working folder: %q", path)

	uiHandler := ui.NewUI(tel)
	uiHandler.Run(func() {
		mainWindow := repoview.NewMainWindow(uiHandler, configService, tel, path)
		uiHandler.OnResizeWindow = mainWindow.OnResizeWindow
		mainWindow.Show()
	})
	tel.SendEvent("program-stop")
	tel.Close()
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
