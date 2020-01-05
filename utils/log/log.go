package log

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils/log/logger"
)

var std = logger.New("")

func Debugf(format string, v ...interface{}) {
	std.Output(logger.Debug, fmt.Sprintf(format, v...))
}

func Infof(format string, v ...interface{}) {
	std.Output(logger.Info, fmt.Sprintf(format, v...))
}

func Warnf(format string, v ...interface{}) {
	std.Output(logger.Warn, fmt.Sprintf(format, v...))
}

func Errorf(format string, v ...interface{}) string {
	msg := fmt.Sprintf(format, v...)
	std.Output(logger.Fatal, msg)
	return msg
}

func Error(v ...interface{}) string {
	msg := fmt.Sprint(v...)
	std.Output(logger.Fatal, msg)
	return msg
}
