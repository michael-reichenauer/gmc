package logger

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/denisbrodbeck/machineid"
)

const (
	Debug = "DEBUG"
	Info  = "INFO "
	Warn  = "WARN "
	Error = "ERROR"
	Fatal = "FATAL"
)

const (
	loggerPathPrefix = "utils/log/logger/logger.go"
	errorsFile       = "gmc_errors.log"
)

var (
	baseFilePathLength = getBaseFileBathLength()
	StdLogger          = NewLogger(fmt.Sprintf("%s:gmc:", getLogID()))
)

const (
	timeFormat = "15:04:05.000"
)

type Logger struct {
	prefix    string // prefix to write at beginning of each line
	isWindows bool
	udpLogger *net.UDPConn
	file      *os.File
}

func RedirectStdErrorToFile() {
	// The error log file in users home dir
	home, err := os.UserHomeDir()
	if err != nil {
		StdLogger.Fatal(err)
	}
	errorsPath := path.Join(home, errorsFile)

	// Log previous error for last instance if it exists
	previousErrorData, _ := ioutil.ReadFile(errorsPath)
	previousError := string(previousErrorData)
	if previousError != "" {
		fileTime := time.Now()
		info, err2 := os.Stat(errorsPath)
		if err2 == nil {
			fileTime = info.ModTime()
		}
		StdLogger.Errorf("Previous instance error at %v:\n%s", fileTime, previousError)
	}
	_ = os.Remove(errorsPath)

	// Redirect std error to error file
	ef, err := os.OpenFile(errorsPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		StdLogger.Fatal(err)
	}
	redirectStdErrToFile(ef)
}

func NewLogger(prefix string) *Logger {
	// fmt.Printf("log target: %s\n", logTarget)
	remoteAddr, err := net.ResolveUDPAddr("udp", "192.168.0.9:40000")
	if err != nil {
		panic(err)
	}
	localAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:0")
	if err != nil {
		panic(err)
	}
	udp, err := net.DialUDP("udp", localAddr, remoteAddr)
	if err != nil {
		panic(err)
	}

	udpLogger := udp
	logFilePath := fmt.Sprintf("%s/gmc.log", os.TempDir())
	_ = os.Remove(logFilePath)
	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}

	return &Logger{prefix: prefix, isWindows: runtime.GOOS == "windows", udpLogger: udpLogger, file: f}
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.Output(Debug, fmt.Sprintf(format, v...))
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.Output(Info, fmt.Sprintf(format, v...))
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.Output(Warn, fmt.Sprintf(format, v...))
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.Output(Error, fmt.Sprintf(format, v...))
}

func (l *Logger) Fatalf(err error, format string, v ...interface{}) string {
	msg := fmt.Sprintf(format, v...)
	emsg := fmt.Sprintf("%s, %v\n%s", msg, err, debug.Stack())
	l.Output(Fatal, emsg)
	l.FatalError(err, msg)
	return emsg
}

func (l *Logger) Fatal(err error, v ...interface{}) string {
	msg := fmt.Sprint(v...)
	emsg := fmt.Sprintf("%s, %v\n%s", msg, err, debug.Stack())

	l.Output(Fatal, emsg)
	l.FatalError(err, msg)
	return emsg
}

func (l *Logger) Output(level string, msg string) {
	l.output(level, msg)
}

func (l *Logger) Outputf(level string, format string, v ...interface{}) {
	l.output(level, fmt.Sprintf(format, v...))
}

func (l *Logger) output(level, message string) {
	//now := time.Now()
	file, line, function := l.getCallerInfo()

	if len(file) > baseFilePathLength {
		file = file[baseFilePathLength:]
	}
	StdTelemetry.SendTrace(level, fmt.Sprintf("%s:%s(%d) %s", file, function, line, message))

	lines := strings.Split(message, "\n")
	for _, ml := range lines {
		txt := fmt.Sprintf("%s%s %s:%s(%d) %s", l.prefix, level, file, function, line, ml)
		_, _ = l.udpLogger.Write([]byte(txt))
		txt2 := fmt.Sprintf("%s %s %s(%d) %s: %s", time.Now().Format(timeFormat), level, file, line, function, ml)
		l.outputToFile(txt2)
	}
}

func (l *Logger) outputToFile(text string) {
	if _, err := l.file.WriteString(text + "\n"); err != nil {
		log.Println(err)
	}
}

func (l *Logger) getCallerInfo() (string, int, string) {
	_, file, line, function, _ := caller(6)
	i := strings.LastIndex(function, ".")
	if i != -1 {
		function = function[i+1:]
	}
	return file, line, function
}

func (l *Logger) FatalError(err error, msg string) {
	StdTelemetry.SendFatalf(err, msg)
}

func getBaseFileBathLength() int {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return 0
	}

	if len(file) > len(loggerPathPrefix) {
		return len(file) - len(loggerPathPrefix)
	}
	return 0
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

func getLogID() string {
	id, err := machineid.ProtectedID("gmc")
	if err != nil {
		return "gmcid"
		//panic(err) ! investigate !!!!!!!!!
	}
	return strings.ToUpper(id[:4])
}
