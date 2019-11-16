package logger

import (
	"syscall"
	"unsafe"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32")
	outputDebugStringW = kernel32.NewProc("OutputDebugStringW")
)

func print(text string) {
	textPtr, err := syscall.UTF16PtrFromString(text)
	if err == nil {
		outputDebugStringW.Call(uintptr(unsafe.Pointer(textPtr)))
	}
}
