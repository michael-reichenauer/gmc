package main

import (
	. "github.com/michael-reichenauer/gmc/build/utils"
	"os"
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
	Must(Cmd("gmc.exe", "-version"))
	Echo("")
	Echo("")

	Must(OpenBrowser("https://github.com/michael-reichenauer/gmc/releases/new"))
}
