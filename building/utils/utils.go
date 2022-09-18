package tools

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"runtime"
	"time"
)

var Root = getRoot()

func EchoLn() {
	fmt.Printf("\n")
}

func Echo(format string, args ...interface{}) {
	fmt.Printf(Now() + " " + format+"\n", args...)
}

func ErrorEcho(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, Now() + " " + format+"\n", args...)
}

func Cmd(cmd string, args ...string) error {
	return CmdWithEnv(nil, cmd, args...)
}

func Now() string{
    return time.Now().Format("2006-01-02T15:04:05.0000")
}

// https://blog.kowalczyk.info/article/wOYk/advanced-command-execution-in-go-with-osexec.html
func CmdWithEnv(env []string, cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	c.Dir = Root
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Env = os.Environ()
	c.Env = append(c.Env, env...)
	return c.Run()
}

func Path(elem ...string) string {
	return path.Join(append([]string{Root}, elem...)...)
}

func RemoveFile(path string) error {
	if err := os.Remove(path); !os.IsNotExist(err) {
		return err
	}

	return nil
}

func CopyFile(src, dst string) error {
    bytesRead, err := ioutil.ReadFile(src)
    if err != nil {
        return err
    }

    err = ioutil.WriteFile(dst, bytesRead, 0755)
    if err != nil {
		return err
    }

	return nil
}

func Must(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		ErrorEcho("Error: %v\nat:\n  %s:%d", err, file, line)
		EchoLn()
		os.Exit(1)
	}
}

func OpenBrowser(url string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	}
	return fmt.Errorf("unsupported platform")
}

func getRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return path.Dir(path.Dir(path.Dir(file)))
}
