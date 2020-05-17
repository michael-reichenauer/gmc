package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
)

var Root = getRoot()

func Echo(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func ErrorEcho(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func Cmd(cmd string, args ...string) error {
	return CmdWithEnv(nil, cmd, args...)
}

func CmdWithEnv(env []string, cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	c.Dir = Root
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Env = os.Environ()
	c.Env = append(c.Env, env...)

	err := c.Run()
	if err != nil {
		errorText := ""
		if ee, ok := err.(*exec.ExitError); ok {
			errorText = string(ee.Stderr)
			errorText = strings.ReplaceAll(errorText, "\t", "   ")
		}
		err := fmt.Errorf("failed: %v\n%v", err, errorText)
		return err
	}
	return nil
}

func Path(elem ...string) string {
	return path.Join(append([]string{Root}, elem...)...)
}

func RemoveFile(path string) error {
	err := os.Remove(path)
	if !os.IsNotExist(err) {
		return err
	}
	return nil
}

func Must(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		ErrorEcho("Error: %v\nat:\n  %s:%d", err, file, line)
		ErrorEcho("")
		os.Exit(1)
	}
}

func getRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return path.Dir(path.Dir(path.Dir(file)))
}

// func copyAndCapture(w io.Writer, r io.Reader) ([]byte, error) {
// 	var out []byte
// 	buf := make([]byte, 1024, 1024)
// 	for {
// 		n, err := r.Read(buf[:])
// 		if n > 0 {
// 			d := buf[:n]
// 			out = append(out, d...)
// 			_, err := w.Write(d)
// 			if err != nil {
// 				return out, err
// 			}
// 		}
// 		if err != nil {
// 			// Read returns io.EOF at the end of file, which is not an error for us
// 			if err == io.EOF {
// 				err = nil
// 			}
// 			return out, err
// 		}
// 	}
// }

// func cmd(cmd string, args ...string) (string, error) {
// 	c := exec.Command(cmd, args...)
// 	c.Dir = root
// 	var stdout []byte
// 	var errStdout, errStderr error
// 	stdoutIn, _ := c.StdoutPipe()
// 	stderrIn, _ := c.StderrPipe()
// 	err := c.Start()
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	// cmd.Wait() should be called only after we finish reading
// 	// from stdoutIn and stderrIn.
// 	// wg ensures that we finish
// 	var wg sync.WaitGroup
// 	wg.Add(1)
// 	go func() {
// 		stdout, errStdout = copyAndCapture(os.Stdout, stdoutIn)
// 		wg.Done()
// 	}()
//
// 	_, errStderr = copyAndCapture(os.Stderr, stderrIn)
//
// 	wg.Wait()
//
// 	err = c.Wait()
// 	if err != nil {
// 		return string(stdout), err
// 	}
// 	if errStdout != nil || errStderr != nil {
// 		return string(stdout), fmt.Errorf("failed to capture stdout or stderr")
// 	}
//
// 	outStr := string(stdout)
// 	return outStr, nil
// }

// func cmd(cmd string, args ...string) (string, error) {
// 	argsText := strings.Join(args, " ")
// 	c := exec.Command(cmd, args...)
// 	c.Dir = root
//
//
// 	out, err := c.Output()
// 	if err != nil {
// 		errorText := ""
// 		if ee, ok := err.(*exec.ExitError); ok {
// 			errorText = string(ee.Stderr)
// 			errorText = strings.ReplaceAll(errorText, "\t", "   ")
// 		}
// 		err := fmt.Errorf("failed: %s %s\n%v\n%v", cmd, argsText, err, errorText)
// 		return string(out), err
// 	}
// 	//log.Infof("OK: git %s %v", argsText, st)
// 	return string(out), nil
// }
