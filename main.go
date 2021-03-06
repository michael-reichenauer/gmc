package main

import (
	"flag"
	"fmt"
	"github.com/michael-reichenauer/gmc/client/console"
	"github.com/michael-reichenauer/gmc/server"
	"github.com/michael-reichenauer/gmc/utils/rpc"
	"io/ioutil"
	stdlog "log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/exec"
	"runtime"

	"github.com/michael-reichenauer/gmc/common/config"
	"github.com/michael-reichenauer/gmc/installation"
	"github.com/michael-reichenauer/gmc/program"
	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/cui"
	"github.com/michael-reichenauer/gmc/utils/git"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/log/logger"
	"github.com/michael-reichenauer/gmc/utils/timer"
)

const (
	version = "0.31"
)

var (
	workingDirFlag  = flag.String("d", "", "specify working directory")
	showVersionFlag = flag.Bool("version", false, "print gmc version")
	pauseFlag       = flag.Bool("pause", false, "pause until user click enter")
	externalWindow  = flag.Bool("external", false, "start gmc in external window (used by ide)")
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

	// if *pauseFlag {
	// 	// The process was started with 'pause' flag, e.g. from Goland,
	// 	// wait until enter is clicked, this can be used for attaching a debugger
	// 	fmt.Printf("Attach a debugger and click 'enter' to proceed ...\n")
	// 	utils.ReadLine()
	// }

	// Disable standard logging since some modules log to stderr, which conflicts with console ui
	stdlog.SetOutput(ioutil.Discard)

	// Set default http client proxy to the system proxy (used by e.g. telemetry)
	utils.SetDefaultHTTPProxy()

	// Enable telemetry and logging
	logger.StdTelemetry.Enable(version)
	defer logger.StdTelemetry.Close()
	log.Eventf("program-start", "Starting gmc %s ...", version)
	logger.RedirectStdErrorToFile()
	logProgramInfo()

	configService := config.NewConfig(version, "")
	configService.SetState(func(s *config.State) {
		s.InstalledVersion = version
	})

	autoUpdate := installation.NewAutoUpdate(configService, version)
	autoUpdate.Start()
	go func() {
		// fs := http.FileServer(http.Dir("C:\\code\\gmc\\client\\wui\\build"))
		// log.Fatal(http.ListenAndServe("127.0.0.1:8081", fs))
		// mape our `/ws` endpoint to the `serveWs` function
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Simple Server")
		})

		pool := server.NewPool()
		go pool.Start()

		http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			server.ServeWs(pool, w, r)
		})
		panic(log.Fatal(http.ListenAndServe(":8080", nil)))
	}()

	// Start rpc sever and serve rpc requests
	rpcServer := rpc.NewServer()
	if err := rpcServer.RegisterService("", server.NewServer(configService)); err != nil {
		panic(log.Fatal(err))
	}
	if err := rpcServer.Start("http://127.0.0.1:9090/api/ws", "/api/events"); err != nil {
		panic(log.Fatal(err))
	}
	defer rpcServer.Close()
	go func() {
		if err := rpcServer.Serve(); err != nil {
			panic(log.Fatal(err))
		}
	}()

	ui := cui.NewCommandUI()
	ui.Run(func() {
		log.Infof("Show main window %s", st)
		mainWindow := console.NewMainWindow(ui)
		mainWindow.Show(rpcServer.URL, *workingDirFlag)
	})
}

func isWindowsGolandConsole() bool {
	return runtime.GOOS == "windows" && *externalWindow && !*pauseFlag
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

func logProgramInfo() {
	log.Infof("Version: %s", version)
	log.Infof("Build: release=%v", program.IsRelease)
	log.Infof("Binary path: %q", utils.BinPath())
	log.Infof("Args: %v", os.Args)
	log.Infof("ID: %q", utils.MachineID)
	log.Infof("OS: %q", runtime.GOOS)
	log.Infof("Arch: %q", runtime.GOARCH)
	log.Infof("Go version: %q", runtime.Version())
	log.Infof("Specified folder path: %q", *workingDirFlag)
	log.Infof("Working Folder: %q", utils.CurrentDir())
	log.Infof("Http proxy: %q", utils.GetHTTPProxyURL())
	go func() { log.Infof("Git version: %q", git.Version()) }()
}
