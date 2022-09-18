package main

import (
	"os"
	"runtime"

	//lint:ignore ST1001 dot import is ok here to increase readability
	. "github.com/michael-reichenauer/gmc/building/utils"
)

func main() {
	os.Chdir(Root)

	Echo("Cleaning ...")
	Must(RemoveFile("gmc.exe"))
	Must(RemoveFile("gmc_linux"))
	Must(RemoveFile("gmc"))
	Must(RemoveFile("gmc_mac"))
	EchoLn()

	Echo("Running tests ...")
	Must(Cmd("go", "test", "-count=1", "./..."))
	EchoLn()

	// linker flags "-ldflags", "-s -w" omits debug and symbols to reduce size
	// ("-H=windowsgui" would disable console on windows)
	Echo("Building Windows gmc.exe ...")
	env := []string{"GOOS=windows", "GOARCH=amd64"}
	Must(CmdWithEnv(env, "go", "build", "-tags", "release", "-ldflags", "-s -w", "-o", "gmc.exe", "main.go"))

	Echo("Building Linux gmc ...")
	env = []string{"GOOS=linux", "GOARCH=amd64"}
	Must(CmdWithEnv(env, "go", "build", "-tags", "release", "-ldflags", "-s -w", "-o", "gmc_linux", "main.go"))
	Must(CopyFile("gmc_linux", "gmc"))

	// Echo("Building Mac gmc_mac ...")
	// env = []string{"GOOS=darwin", "GOARCH=amd64"}
	// Must(CmdWithEnv(env, "go", "build", "-tags", "release", "-ldflags", "-s -w", "-o", "gmc_mac", "main.go"))
	// EchoLn()

	Echo("Built version:")
	echoBuiltVersion()
	EchoLn()

	Echo("Publish url: https://github.com/michael-reichenauer/gmc/releases/new")
	_ = OpenBrowser("https://github.com/michael-reichenauer/gmc/releases/new")

	EchoLn()
	EchoLn()
}

func echoBuiltVersion() {
	switch runtime.GOOS {
	case "linux":
		Must(Cmd("./gmc", "-version"))
	case "windows":
		Must(Cmd("./gmc.exe", "-version"))
	case "darwin":
		Must(Cmd("./gmc_mac", "-version"))
	}
}
