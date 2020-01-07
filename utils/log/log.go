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

func Fatalf(err error, format string, v ...interface{}) string {
	msg := fmt.Sprintf(format, v...)
	emsg := fmt.Sprintf("%s, %v", msg, err)
	logger.Std.Output(logger.Fatal, emsg)
	logger.Std.FatalError(err, msg)
	return emsg
}

func Fatal(err error, v ...interface{}) string {
	msg := fmt.Sprint(v...)
	emsg := err.Error()
	if len(v) > 0 {
		emsg = fmt.Sprintf("%s, %v", msg, err)
	}
	logger.Std.Output(logger.Fatal, emsg)
	logger.Std.FatalError(err, msg)
	return emsg
}
