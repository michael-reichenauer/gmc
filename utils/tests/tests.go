package tests

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
)

type TempFolder string
type TempFile string

func CreateTempFolder() TempFolder {
	basePath := TempBasePath()

	err := os.Mkdir(basePath, 0700)
	if err != nil && !os.IsExist(err) {
		panic(log.Fatal(err))
	}

	dir, err := os.MkdirTemp(basePath, "folder")
	if err != nil {
		panic(log.Fatal(err))
	}
	dir = strings.ReplaceAll(dir, "\\", "/")
	return TempFolder(dir)
}

func GetTempPath() string {
	basePath := TempBasePath()

	err := os.Mkdir(basePath, 0700)
	if err != nil && !os.IsExist(err) {
		panic(log.Fatal(err))
	}

	dir, err := ioutil.TempDir(basePath, "folder")
	if err != nil {
		panic(log.Fatal(err))
	}
	dir = strings.ReplaceAll(dir, "\\", "/")

	_ = os.RemoveAll(dir)

	return dir
}

func (t TempFolder) Path(elem ...string) string {
	return path.Join(append([]string{string(t)}, elem...)...)
}

func (t TempFolder) MkDir(elem ...string) TempFolder {
	path := t.Path(elem...)
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		panic(log.Fatal(err))
	}
	return TempFolder(path)
}

func (t TempFolder) File(elem ...string) TempFile {
	return TempFile(t.Path(elem...))
}

func (t TempFile) Write(text string) {
	if err := utils.FileWrite(string(t), []byte(text)); err != nil {
		panic(log.Fatal(err))
	}
}
func (t TempFile) Read() string {
	text := string(utils.MustFileRead(string(t)))
	text = strings.ReplaceAll(text, "\r", "")
	return text
}
func (t TempFile) TryRead() (string, error) {
	f, err := utils.FileRead(string(t))
	if err != nil {
		return "", err
	}
	return string(f), err
}

func CleanTemp() {
	_ = os.RemoveAll(TempBasePath())
}

func TempBasePath() string {
	tmpDir := strings.ReplaceAll(os.TempDir(), "\\", "/")
	return path.Join(tmpDir, "gmctmp")
}

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
