package log

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils/log/logger"
)

func Debugf(format string, v ...interface{}) {
	logger.StdLogger.Debugf(format, v...)
}

func Infof(format string, v ...interface{}) {
	logger.StdLogger.Infof(format, v...)
}

func Warnf(format string, v ...interface{}) {
	logger.StdLogger.Warnf(format, v...)
}

func Fatalf(err error, format string, v ...interface{}) string {
	return logger.StdLogger.Fatalf(err, format, v...)
}

func Fatal(err error, v ...interface{}) string {
	return logger.StdLogger.Fatal(err, v...)
}

func Event(eventName string) {
	logger.StdLogger.Infof("Telemetry: event: %q", eventName)
	logger.StdTelemetry.SendEvent(eventName)
}

func Eventf(eventName, message string, v ...interface{}) {
	logger.StdLogger.Infof("Telemetry: event: %q", fmt.Sprintf(message, v...))
	logger.StdTelemetry.SendEventf(eventName, message, v...)
}
