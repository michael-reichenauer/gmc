package logger

import (
	"fmt"
	"net"
	"runtime"
	"strings"
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
)

var (
	baseFilePathLength = getBaseFileBathLength()
	StdLogger          = NewLogger("gmc:")
)

type Logger struct {
	prefix    string // prefix to write at beginning of each line
	isWindows bool
	udpLogger *net.UDPConn
}

func NewLogger(prefix string) *Logger {
	// fmt.Printf("log target: %s\n", logTarget)
	remoteAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:6868")
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

	return &Logger{prefix: prefix, isWindows: runtime.GOOS == "windows", udpLogger: udpLogger}
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

func (l *Logger) Fatalf(err error, format string, v ...interface{}) string {
	msg := fmt.Sprintf(format, v...)
	emsg := fmt.Sprintf("%s, %v", msg, err)
	l.Output(Fatal, emsg)
	l.FatalError(err, msg)
	return emsg
}

func (l *Logger) Fatal(err error, v ...interface{}) string {
	msg := fmt.Sprint(v...)
	emsg := err.Error()
	if len(v) > 0 {
		emsg = fmt.Sprintf("%s, %v", msg, err)
	}
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
	file, line := l.getCallerInfo()

	if len(file) > baseFilePathLength {
		file = file[baseFilePathLength:]
	}
	StdTelemetry.SendTrace(level, fmt.Sprintf("%s(%d) %s", file, line, message))

	lines := strings.Split(message, "\n")
	for _, ml := range lines {
		txt := fmt.Sprintf("%s%s %s(%d) %s", l.prefix, level, file, line, ml)
		//print(txt)
		_, err := l.udpLogger.Write([]byte(txt))
		if err != nil {
			fmt.Printf("Failed to log %v\n", err)
		}
	}
}

func (l *Logger) getCallerInfo() (string, int) {
	_, file, line, _ := runtime.Caller(5)
	return file, line
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
