// +build !windows

package logger

import (
	stdlog "log"
	"os"
	"syscall"
)

func print(text string) {
	// Nothing yet
}

// redirectStderr to the file passed in
func redirectStderr(f *os.File) {
	err := syscall.Dup2(int(f.Fd()), int(os.Stderr.Fd()))
	if err != nil {
		stdlog.Fatalf("Failed to redirect stderr to file: %v", err)
	}
}
