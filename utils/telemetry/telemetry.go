package telemetry

import (
	"fmt"
	"github.com/Microsoft/ApplicationInsights-Go/appinsights"
	"github.com/Microsoft/ApplicationInsights-Go/appinsights/contracts"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/log/logger"
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
	startTime = time.Now()
)

type Telemetry struct {
	version string
	client  appinsights.TelemetryClient
	limit   *rate.Limiter
}

func NewTelemetry(version string) *Telemetry {
	h := &Telemetry{version: version, limit: rate.NewLimiter(1, 300)}
	h.client = h.createClient()
	return h
}

func (h *Telemetry) SendTrace(level, text string) {
	switch level {
	case logger.Info:
		h.send(appinsights.NewTraceTelemetry(text, contracts.Information))
	case logger.Warn:
		h.send(appinsights.NewTraceTelemetry(text, contracts.Warning))
	case logger.Error:
		h.send(appinsights.NewTraceTelemetry(text, contracts.Error))
	}
}

func (h *Telemetry) SendEvent(eventName string) {
	log.Infof("Send event: %q", eventName)
	h.send(appinsights.NewEventTelemetry(eventName))
}

func (h *Telemetry) SendEventf(eventName, message string, v ...interface{}) {
	event := appinsights.NewEventTelemetry(eventName)
	msg := fmt.Sprintf(message, v...)
	log.Infof("Send event: %q, %q", eventName, msg)
	event.Properties["Message"] = msg
	h.send(event)
}

func (h *Telemetry) SendError(err error) {
	log.Warnf("Send error: %v", err)
	if h.send(appinsights.NewExceptionTelemetry(err)) {
		h.client.Channel().Flush()
	}
}

func (h *Telemetry) SendErrorf(err error, message string, v ...interface{}) {
	t := appinsights.NewExceptionTelemetry(err)
	msg := fmt.Sprintf(message, v...)
	t.Properties["Message"] = msg
	log.Warnf("Send error: %q, %v", msg, err)
	if h.send(t) {
		h.client.Channel().Flush()
	}
}

func (h *Telemetry) Close() {
	log.Infof("Close telemetry")
	select {
	case <-h.client.Channel().Close(10 * time.Second):
		// Ten second timeout for retries.
	case <-time.After(20 * time.Second):
		// Absolute timeout.
	}
}

func (h *Telemetry) createClient() appinsights.TelemetryClient {
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
	client.Context().Tags.Application().SetVer(h.version)
	h.setCommonProperties(client.Context().CommonProperties)

	// appinsights.NewDiagnosticsMessageListener(func(msg string) error {
	// 	log.Debugf("Telemetry: %s", msg)
	// 	return nil
	// })

	return client
}

func (h *Telemetry) send(telemetry appinsights.Telemetry) bool {
	if !h.limit.Allow() {
		return false
	}

	h.client.Track(telemetry)
	return true
}

func (h *Telemetry) setCommonProperties(properties map[string]string) {
	properties["User"] = h.getUserID()
}

func (h *Telemetry) getUserID() string {
	u, err := user.Current()
	if err != nil {
		return h.getMachineID()
	}
	return u.Username + "_" + h.getMachineID()
}

func (h *Telemetry) getMachineID() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		panic(log.Errorf("Could not get interfaces, %v", err))
	}
	for _, ifx := range interfaces {
		return strings.ToUpper(strings.Replace(ifx.HardwareAddr.String(), ":", "", -1))
	}
	return "000000000000"
}
