package logger

import (
	"fmt"
	"github.com/Microsoft/ApplicationInsights-Go/appinsights/contracts"
	"github.com/microsoft/ApplicationInsights-Go/appinsights"
	"golang.org/x/time/rate"
	"net"
	"os/user"
	"strings"
	"time"
)

const (
	instrumentationKeyProd = ""
	instrumentationKeyDev  = "800c3ff8-ecdc-435c-adfd-8a511c688cd9"
)

var (
	startTime    = time.Now()
	StdTelemetry = NewTelemetry()
)

type telemetry struct {
	client appinsights.TelemetryClient
	limit  *rate.Limiter
}

func NewTelemetry() *telemetry {
	h := &telemetry{limit: rate.NewLimiter(1, 300)}
	return h
}

func (h *telemetry) Enable(version string) {
	h.client = h.createClient(version)
}

func (h *telemetry) SendTrace(level, text string) {
	if h.client == nil {
		// Not enabled
		return
	}
	switch level {
	case Info:
		h.send(appinsights.NewTraceTelemetry(text, contracts.Information))
	case Warn:
		h.send(appinsights.NewTraceTelemetry(text, contracts.Warning))
	case Error:
		h.send(appinsights.NewTraceTelemetry(text, contracts.Error))
		h.client.Channel().Flush()
	}
}

func (h *telemetry) SendEvent(eventName string) {
	if h.client == nil {
		// Not enabled
		return
	}
	h.send(appinsights.NewEventTelemetry(eventName))
}

func (h *telemetry) SendEventf(eventName, message string, v ...interface{}) {
	if h.client == nil {
		// Not enabled
		return
	}
	event := appinsights.NewEventTelemetry(eventName)
	msg := fmt.Sprintf(message, v...)
	event.Properties["Message"] = msg
	h.send(event)
}

func (h *telemetry) SendError(err error) {
	if h.client == nil {
		// Not enabled
		return
	}
	StdLogger.Warnf("Send error: %v", err)
	if h.send(appinsights.NewExceptionTelemetry(err)) {
		h.client.Channel().Flush()
	}
}
func (h *telemetry) SendFatalf(err error, message string, v ...interface{}) {
	if h.client == nil {
		// Not enabled
		return
	}
	StdLogger.Warnf("Send fatal: %v", err)
	t := appinsights.NewExceptionTelemetry(err)
	t.Frames = appinsights.GetCallstack(4)
	msg := fmt.Sprintf(message, v...)
	t.Properties["Message"] = msg
	StdLogger.Warnf("Send error: %q, %v", msg, err)
	h.send(t)
	h.Close()
}

func (h *telemetry) SendErrorf(err error, message string, v ...interface{}) {
	if h.client == nil {
		// Not enabled
		return
	}
	t := appinsights.NewExceptionTelemetry(err)
	msg := fmt.Sprintf(message, v...)
	t.Properties["Message"] = msg
	StdLogger.Warnf("Send error: %q, %v", msg, err)
	if h.send(t) {
		h.client.Channel().Flush()
	}
}

func (h *telemetry) Close() {
	if h.client == nil {
		// Not enabled
		return
	}
	StdLogger.Infof("Close telemetry")
	select {
	case <-h.client.Channel().Close(10 * time.Second):
		// Ten second timeout for retries.
	case <-time.After(20 * time.Second):
		// Absolute timeout.
	}
}

func (h *telemetry) createClient(version string) appinsights.TelemetryClient {
	var instrumentationKey string
	// if isProduction {
	// 	instrumentationKey = instrumentationKeyProd
	// } else {
	instrumentationKey = instrumentationKeyDev
	//}

	config := appinsights.NewTelemetryConfiguration(instrumentationKey)
	client := appinsights.NewTelemetryClientFromConfig(config)
	client.Context().Tags.User().SetId(h.getUserID())
	client.Context().Tags.Cloud().SetRoleInstance(h.getMachineID())
	client.Context().Tags.Session().SetId(startTime.String())
	client.Context().Tags.Application().SetVer(version)
	h.setCommonProperties(client.Context().CommonProperties)

	// appinsights.NewDiagnosticsMessageListener(func(msg string) error {
	// 	log.Debugf("Telemetry: %s", msg)
	// 	return nil
	// })

	return client
}

func (h *telemetry) send(telemetry appinsights.Telemetry) bool {
	if !h.limit.Allow() {
		return false
	}

	h.client.Track(telemetry)
	return true
}

func (h *telemetry) setCommonProperties(properties map[string]string) {
	// properties["User"] = h.getUserID()
}

func (h *telemetry) getUserID() string {
	u, err := user.Current()
	if err != nil {
		return h.getMachineID()
	}
	return u.Username + "_" + h.getMachineID()
}

func (h *telemetry) getMachineID() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		panic(StdLogger.Fatalf(err, "Could not get interfaces"))
	}
	for _, ifx := range interfaces {
		return strings.ToUpper(strings.Replace(ifx.HardwareAddr.String(), ":", "", -1))
	}
	return "000000000000"
}
