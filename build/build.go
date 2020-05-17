package main

import (
	. "github.com/michael-reichenauer/gmc/build/utils"
	"os"
	"runtime"
)

func main() {
	os.Chdir(Root)

	Echo("Cleaning ...")
	Must(RemoveFile("gmc.exe"))
	Must(RemoveFile("gmc_linux"))
	Must(RemoveFile("gmc_mac"))
	Echo("")

	Echo("Running tests ...")
	Must(Cmd("go", "test", "-count=1", "./..."))
	Echo("")

	Echo("Building gmc.exe ...")
	env := []string{"GOOS=windows", "GOARCH=amd64"}
	Must(CmdWithEnv(env, "go", "build", "-tags", "release", "-ldflags", "-s -w", "-o", "gmc.exe", "main.go"))

	Echo("Building gmc_linux ...")
	env = []string{"GOOS=linux", "GOARCH=amd64"}
	Must(CmdWithEnv(env, "go", "build", "-tags", "release", "-ldflags", "-s -w", "-o", "gmc_linux", "main.go"))

	Echo("Building gmc_mac ...")
	env = []string{"GOOS=darwin", "GOARCH=amd64"}
	Must(CmdWithEnv(env, "go", "build", "-tags", "release", "-ldflags", "-s -w", "-o", "gmc_mac", "main.go"))
	Echo("")

	Echo("Built version:")
	switch runtime.GOOS {
	case "linux":
		Must(Cmd("./gmc_linux", "-version"))
	case "windows":
		Must(Cmd("./gmc.exe", "-version"))
		Must(OpenBrowser("https://github.com/michael-reichenauer/gmc/releases/new"))
	case "darwin":
		Must(Cmd("./gmc_mac", "-version"))
	}
	Echo("")
	Echo("")
}
