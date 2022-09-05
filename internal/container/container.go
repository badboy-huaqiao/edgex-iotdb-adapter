package container

import (
	"github.com/edgexfoundry/edgexfoundry-iotdb-connector/internal/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-messaging/v2/messaging"
)

// ConfigurationName contains the name of command's config.ConfigurationStruct implementation in the DIC.
var ConfigurationName = di.TypeInstanceToName(config.ConfigurationStruct{})

// ConfigurationFrom helper function queries the DIC and returns command's config.ConfigurationStruct implementation.
func ConfigurationFrom(get di.Get) *config.ConfigurationStruct {
	return get(ConfigurationName).(*config.ConfigurationStruct)
}

// MessagingClientName contains the name of the messaging client instance in the DIC.
var MessagingClientName = di.TypeInstanceToName((*messaging.MessageClient)(nil))

// MessagingClientFrom helper function queries the DIC and returns the messaging client.
func MessagingClientFrom(get di.Get) messaging.MessageClient {
	client, ok := get(MessagingClientName).(messaging.MessageClient)
	if !ok {
		return nil
	}

	return client
}
