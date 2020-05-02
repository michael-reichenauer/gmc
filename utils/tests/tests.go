package tests

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils/log"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"
)

func ManualTest(t *testing.T) {
	t.Helper()
	_, _, _, function, _ := caller(2)
	i := strings.LastIndex(function, ".")
	if i == -1 {
		return
	}
	name := function[i+1:]
	path := function[:i]

	p := strings.ReplaceAll(path, "/", "_")
	p = strings.ReplaceAll(p, ".", "_")
	p = strings.ReplaceAll(p, "-", "_")
	if strings.Contains(os.Args[0], p) &&
		stringsContains(os.Args, fmt.Sprintf("^%s$", name)) {
		// The test runner is specified to run test for that function name
		return
	}
	t.SkipNow()
}

func caller(skip int) (pc uintptr, file string, line int, function string, ok bool) {
	rpc := make([]uintptr, 1)
	n := runtime.Callers(skip+1, rpc[:])
	if n < 1 {
		return
	}
	frame, _ := runtime.CallersFrames(rpc).Next()
	return frame.PC, frame.File, frame.Line, frame.Function, frame.PC != 0
}

func stringsContains(s []string, e string) bool {
	return stringsIndex(s, e) != -1
}

func stringsIndex(s []string, e string) int {
	for i, a := range s {
		if a == e {
			return i
		}
	}
	return -1
}

func CreateTempFolder() string {
	basePath := TempBasePath()

	err := os.Mkdir(basePath, 0700)
	if err != nil && !os.IsExist(err) {
		panic(log.Fatal(err))
	}

	dir, err := ioutil.TempDir(basePath, "folder")
	if err != nil {
		panic(log.Fatal(err))
	}
	return dir
}

func CleanTemp() {
	_ = os.RemoveAll(TempBasePath())
}

func TempBasePath() string {
	return path.Join(os.TempDir(), "gmctmp")
}
