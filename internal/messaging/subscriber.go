package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/edgexfoundry/edgexfoundry-iotdb-connector/internal/adapter"
	"github.com/edgexfoundry/edgexfoundry-iotdb-connector/internal/container"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	bootstrapMessaging "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-messaging/v2/messaging"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

func BootstrapHandler(ctx context.Context, wg *sync.WaitGroup, startupTimer startup.Timer, dic *di.Container) (success bool) {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)
	messageBusInfo := config.MessageQueue

	messageBusInfo.AuthMode = strings.ToLower(strings.TrimSpace(messageBusInfo.AuthMode))
	if len(messageBusInfo.AuthMode) > 0 && messageBusInfo.AuthMode != bootstrapMessaging.AuthModeNone {
		if err := bootstrapMessaging.SetOptionsAuthData(&messageBusInfo, lc, dic); err != nil {
			lc.Error(err.Error())
			return false
		}
	}

	msgClient, err := messaging.NewMessageClient(
		types.MessageBusConfig{
			PublishHost: types.HostInfo{
				Host:     messageBusInfo.Host,
				Port:     messageBusInfo.Port,
				Protocol: messageBusInfo.Protocol,
			},
			SubscribeHost: types.HostInfo{
				Host:     messageBusInfo.Host,
				Port:     messageBusInfo.Port,
				Protocol: messageBusInfo.Protocol,
			},
			Type:     messageBusInfo.Type,
			Optional: messageBusInfo.Optional,
		})
	if err != nil {
		lc.Errorf("Failed to create MessageClient: %v", err)
		return false
	}
	for startupTimer.HasNotElapsed() {
		select {
		case <-ctx.Done():
			return false
		default:
			err = msgClient.Connect()
			if err != nil {
				lc.Warnf("Unable to connect MessageBus: %w", err)
				startupTimer.SleepForInterval()
				continue
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				<-ctx.Done()
				if msgClient != nil {
					_ = msgClient.Disconnect()
				}
				lc.Infof("Disconnected from MessageBus")
			}()

			dic.Update(di.ServiceConstructorMap{
				container.MessagingClientName: func(get di.Get) interface{} {
					return msgClient
				},
			})

			lc.Info(fmt.Sprintf(
				"Connected to %s Message Bus @ %s://%s:%d publishing on '%s' prefix topic with AuthMode='%s'",
				messageBusInfo.Type,
				messageBusInfo.Protocol,
				messageBusInfo.Host,
				messageBusInfo.Port,
				messageBusInfo.PublishTopicPrefix,
				messageBusInfo.AuthMode))

			return true
		}
	}

	lc.Error("Connecting to MessageBus time out")
	return false
}

func SubscribeEvents(ctx context.Context, dic *di.Container) errors.EdgeX {
	messageBusInfo := container.ConfigurationFrom(dic.Get).MessageQueue
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	messageBus := container.MessagingClientFrom(dic.Get)

	messages := make(chan types.MessageEnvelope)
	messageErrors := make(chan error)

	topics := []types.TopicChannel{
		{
			Topic:    messageBusInfo.SubscribeTopic,
			Messages: messages,
		},
	}

	err := messageBus.Subscribe(topics, messageErrors)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	edgexAdapter := adapter.EdgeXAdapterFrom(dic.Get)

	go func() {
		for {
			select {
			case <-ctx.Done():
				lc.Infof("Exiting waiting for MessageBus '%s' topic messages", messageBusInfo.SubscribeTopic)
				return
			case e := <-messageErrors:
				lc.Error(e.Error())
			case msgEnvelope := <-messages:
				lc.Debugf("Event received on message queue. Topic: %s, Correlation-id: %s ", messageBusInfo.SubscribeTopic, msgEnvelope.CorrelationID)
				event := &requests.AddEventRequest{}
				// decoding the large payload may cause memory issues so checking before decoding
				// maxEventSize := container.ConfigurationFrom(dic.Get).MaxEventSize
				// edgeXerr := utils.CheckPayloadSize(msgEnvelope.Payload, maxEventSize*1000)
				// if edgeXerr != nil {
				// 	lc.Errorf("event size exceed MaxEventSize(%d KB)", maxEventSize)
				// 	break
				// }
				err = unmarshalPayload(msgEnvelope, event)
				if err != nil {
					lc.Errorf("fail to unmarshal event, %v", err)
					break
				}
				err = validateEvent(msgEnvelope.ReceivedTopic, event.Event)
				if err != nil {
					lc.Error(err.Error())
					break
				}
				if err = edgexAdapter.EventWriter(event.Event); err != nil {
					lc.Errorf("fail to write event to iotdb, %v", err)
				}
				// fmt.Printf("Rev event: %v", event)
			}
		}
	}()
	return nil
}

func unmarshalPayload(envelope types.MessageEnvelope, target interface{}) error {
	var err error
	switch envelope.ContentType {
	case common.ContentTypeJSON:
		err = json.Unmarshal(envelope.Payload, target)

	default:
		err = fmt.Errorf("unsupported content-type '%s' recieved", envelope.ContentType)
	}
	return err
}

func validateEvent(messageTopic string, e dtos.Event) errors.EdgeX {
	// Parse messageTopic by the pattern `edgex/events/<device-profile-name>/<device-name>/<source-name>`
	fields := strings.Split(messageTopic, "/")

	// assumes a non-empty base topic with /profileName/deviceName/sourceName appended by publisher
	if len(fields) < 4 {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("invalid message topic %s", messageTopic), nil)
	}

	len := len(fields)
	profileName := fields[len-3]
	deviceName := fields[len-2]
	sourceName := fields[len-1]

	// Check whether the event fields match the message topic
	if e.ProfileName != profileName {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("event's profileName %s mismatches with the name %s received in topic", e.ProfileName, profileName), nil)
	}
	if e.DeviceName != deviceName {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("event's deviceName %s mismatches with the name %s received in topic", e.DeviceName, deviceName), nil)
	}
	if e.SourceName != sourceName {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("event's sourceName %s mismatches with the name %s received in topic", e.SourceName, sourceName), nil)
	}
	return nil
}
