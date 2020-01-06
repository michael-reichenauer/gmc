package log

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils/log/logger"
)

func Debugf(format string, v ...interface{}) {
	logger.Std.Output(logger.Debug, fmt.Sprintf(format, v...))
}

func Infof(format string, v ...interface{}) {
	logger.Std.Output(logger.Info, fmt.Sprintf(format, v...))
}

func Warnf(format string, v ...interface{}) {
	logger.Std.Output(logger.Warn, fmt.Sprintf(format, v...))
}

func Errorf(format string, v ...interface{}) string {
	msg := fmt.Sprintf(format, v...)
	logger.Std.Output(logger.Error, msg)
	return msg
}

func Error(v ...interface{}) string {
	msg := fmt.Sprint(v...)
	logger.Std.Output(logger.Error, msg)
	return msg
}
