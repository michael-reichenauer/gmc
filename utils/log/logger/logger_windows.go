package logger

import (
	stdlog "log"
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32")
	outputDebugStringW = kernel32.NewProc("OutputDebugStringW")
	procSetStdHandle   = kernel32.NewProc("SetStdHandle")
)

func print(text string) {
	textPtr, err := syscall.UTF16PtrFromString(text)
	if err == nil {
		outputDebugStringW.Call(uintptr(unsafe.Pointer(textPtr)))
	}
}

func setStdHandle(stdhandle int32, handle syscall.Handle) error {
	r0, _, e1 := syscall.Syscall(procSetStdHandle.Addr(), 2, uintptr(stdhandle), uintptr(handle), 0)
	if r0 == 0 {
		if e1 != 0 {
			return error(e1)
		}
		return syscall.EINVAL
	}
	return nil
}

// redirectStderr to the file passed in
func redirectStdErrToFile(f *os.File) {
	err := setStdHandle(syscall.STD_ERROR_HANDLE, syscall.Handle(f.Fd()))
	if err != nil {
		stdlog.Fatalf("Failed to redirect stderr to file: %v", err)
	}
	// SetStdHandle does not affect prior references to stderr
	os.Stderr = f
}
