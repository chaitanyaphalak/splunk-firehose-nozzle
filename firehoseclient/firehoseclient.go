package firehoseclient

import (
	"fmt"
	"strings"
	"time"

	"github.com/cloudfoundry-community/firehose-to-syslog/logging"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/nozzle"
	"github.com/gorilla/websocket"
)

type FirehoseNozzle struct {
	consumer     splunknozzle.FirehoseConsumer
	eventRouting splunknozzle.EventRouter
	config       *FirehoseConfig
}

type FirehoseConfig struct {
	TrafficControllerURL   string
	InsecureSSLSkipVerify  bool
	IdleTimeoutSeconds     time.Duration
	FirehoseSubscriptionID string
}

func NewFirehoseNozzle(consumer splunknozzle.FirehoseConsumer, eventRouting splunknozzle.EventRouter, firehoseconfig *FirehoseConfig) *FirehoseNozzle {
	return &FirehoseNozzle{
		eventRouting: eventRouting,
		consumer:     consumer,
		config:       firehoseconfig,
	}
}

func (f *FirehoseNozzle) Start() error {
	return f.routeEvent()
}

func (f *FirehoseNozzle) routeEvent() error {
	messages, errs := f.consumer.Firehose(f.config.FirehoseSubscriptionID, "")
	ticker := time.NewTicker(time.Second * 30)
	msgSeen := int64(0)

	for {
		select {
		case envelope := <-messages:
			if envelope.LogMessage != nil && envelope.LogMessage.Message != nil {
				if strings.Contains(string(envelope.LogMessage.Message), "PR-34") {
					msgSeen += 1
				}
			}
			f.eventRouting.RouteEvent(envelope)
		case err := <-errs:
			f.handleError(err)
			return err
		case <-ticker.C:
			fmt.Printf("Have seen %d from firehose for PR-34\n", msgSeen)
		}
	}

	ticker.Stop()
	return nil
}

func (f *FirehoseNozzle) handleError(err error) {
	switch {
	case websocket.IsCloseError(err, websocket.CloseNormalClosure):
		logging.LogError("Normal Websocket Closure", err)
	case websocket.IsCloseError(err, websocket.ClosePolicyViolation):
		logging.LogError("Error while reading from the firehose", err)
		logging.LogError("Disconnected because nozzle couldn't keep up. Please try scaling up the nozzle.", nil)

	default:
		logging.LogError("Error while reading from the firehose", err)
	}

	logging.LogError("Closing connection with traffic controller due to", err)
	f.consumer.Close()
}
