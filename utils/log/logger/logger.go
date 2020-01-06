package logger

import (
	"fmt"
	"runtime"
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
	Std                = New("")
)

type Logger struct {
	prefix    string // prefix to write at beginning of each line
	isWindows bool
	target    func(level string, msg string)
}

func New(prefix string) *Logger {
	return &Logger{prefix: prefix, isWindows: runtime.GOOS == "windows"}
}

func (l *Logger) SetTarget(target func(level string, msg string)) {
	l.target = target
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
	text := fmt.Sprintf("%s(%d) %s", file, line, message)
	print(fmt.Sprintf("%s%s %s", l.prefix, level, text))
	if l.target != nil {
		l.target(level, text)
	}
}

func (l *Logger) getCallerInfo() (string, int) {
	_, file, line, _ := runtime.Caller(4)
	return file, line
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
