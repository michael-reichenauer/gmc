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

func Fatalf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	std.Output(logger.Fatal, msg)
	panic(msg)
}

func Fatal(v ...interface{}) {
	msg := fmt.Sprint(v...)
	std.Output(logger.Fatal, msg)
	panic(msg)
}

//// Fatal is equivalent to Print() followed by a call to os.Exit(1).
//func Fatal(v ...interface{}) {
//	std.Output(2, fmt.Sprint(v...))
//	os.Exit(1)
//}
//
//// Fatalf is equivalent to Printf() followed by a call to os.Exit(1).
//func Fatalf(format string, v ...interface{}) {
//	std.Output(2, fmt.Sprintf(format, v...))
//	os.Exit(1)
//}
