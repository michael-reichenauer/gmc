package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	stdlog "log"
	_ "net/http/pprof"

	"github.com/michael-reichenauer/gmc/client/console"
	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/installation"
	"github.com/michael-reichenauer/gmc/program"
	"github.com/michael-reichenauer/gmc/server"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/log/logger"
	"github.com/michael-reichenauer/gmc/utils/one"
	"github.com/michael-reichenauer/gmc/utils/rpc"
)

const (
	version = "0.53"
)

var (
	workingDirFlag  = flag.String("d", "", "specify working directory")
	showVersionFlag = flag.Bool("version", false, "print gmc version")
	pauseFlag       = flag.Bool("pause", false, "pause at start until user click enter (allow time to attach debugger)")
	externalWindow  = flag.Bool("external", false, "start gmc in external window (used by ide)")
)

func main() {
	flag.Parse()

	if *showVersionFlag {
		fmt.Printf("%s", version)
		return
	}

	if program.IsWindowsGolandConsole(*externalWindow, *pauseFlag) {
		// Seems program is run within a debugger console, which does not support termbox
		// So a new external process is started and this instance ends
		program.StartAsExternalProcess()
		return
	}

	if *pauseFlag {
		// The process was started with 'pause' flag, e.g. from an IDE,
		// wait until enter is clicked, this can be used for attaching a debugger
		fmt.Printf("Attach a debugger and click 'enter' to proceed ...\n")
		utils.ReadLine()
	}

	// Disable standard logging since some modules log to stderr, which conflicts with console ui
	stdlog.SetOutput(ioutil.Discard)

	// Set default http client proxy to the system proxy (used by e.g. autoUpdate and telemetry)
	utils.SetDefaultHTTPProxy()

	// Enable telemetry and logging
	// logger.StdTelemetry.Enable(version)
	// defer logger.StdTelemetry.Close()
	log.Eventf("program-start", "Starting gmc %s ...", version)

	// Redirect StdError to file to handle panic output.
	// Next run will log error file with the previous panic output.
	logger.RedirectStdErrorToFile()

	program.LogProgramInfo(version, *workingDirFlag)

	configService := config.NewConfig(version, "")
	configService.SetState(func(s *config.State) {
		s.InstalledVersion = version
	})

	autoUpdate := installation.NewAutoUpdate(configService, version)
	autoUpdate.Start()

	// Start rpc sever and serve rpc requests
	rpcServer := rpc.NewServer()
	if err := rpcServer.RegisterService("", server.NewApiServer(configService)); err != nil {
		panic(log.Fatal(err))
	}
	if err := rpcServer.Start("http://127.0.0.1:0/api/ws", "/api/events"); err != nil {
		panic(log.Fatal(err))
	}
	defer rpcServer.Close()

	go func() {
		if err := rpcServer.Serve(); err != nil {
			panic(log.Fatal(err))
		}
	}()

	// Start client cmd ui
	ui := cui.NewCommandUI(version)
	ui.Run(func() {
		one.RunWith(func() {
			mainWindow := console.NewMainWindow(ui)
			mainWindow.Show(rpcServer.URL, *workingDirFlag)
		}, func(f func()) {
			ui.Post(f)
		})
	})
}
