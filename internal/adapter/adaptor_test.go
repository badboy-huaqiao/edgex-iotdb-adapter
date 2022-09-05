package adapter

import (
	"testing"

	"github.com/apache/iotdb-client-go/client"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
)

var testConfig = AdaptorConfig{
	StorageGroup: defaultEdgeXStorageGroup,
	ClientConfig: client.Config{
		Host:     "192.168.56.13",
		Port:     "6667",
		UserName: "root",
		Password: "root",
	},
}

var testUnsupportedEventType = dtos.Event{
	DeviceName: "Random-UnsignedInteger-Device",
	Origin:     1653207774557490000,
	Readings: []dtos.BaseReading{
		{
			Origin:       1653207774557490000,
			DeviceName:   "Random-UnsignedInteger-Device",
			ResourceName: "Uint64",
			ValueType:    "Uint64",
			SimpleReading: dtos.SimpleReading{
				Value: "9890262310393245329",
			},
		},
	},
}

var testEvent = dtos.Event{
	DeviceName: "Random-Integer-Device",
	Origin:     1653207774557490000,
	Readings: []dtos.BaseReading{
		{
			Origin:       1653207774557490000,
			DeviceName:   "RandomIntegerDevice",
			ResourceName: "Int32",
			ValueType:    "Int32",
			SimpleReading: dtos.SimpleReading{
				Value: "1616959848",
			},
		},
		{
			Origin:       1653207774557490000,
			DeviceName:   "RandomIntegerDevice",
			ResourceName: "Int32",
			ValueType:    "Int32",
			SimpleReading: dtos.SimpleReading{
				Value: "1616959848",
			},
		},
	},
}

func TestInitializer(t *testing.T) {
	adaptor := NewAdaptorWithConfig(testConfig)
	defer func() {
		if adaptor.session != nil {
			adaptor.session.Close()
			t.Log("session is not null, close it.")
		}
	}()
	if err := adaptor.Initializer(nil); err != nil {
		t.Errorf("Initializer failed, err: %s\n", err.Error())
		t.Errorf("Initializer failed, can't connect to iotdb with: [host=%s,port=%s]\n", testConfig.ClientConfig.Host, testConfig.ClientConfig.Port)
	}
}
func TestEventWriter(t *testing.T) {
	adaptor := NewAdaptorWithConfig(testConfig)
	defer func() {
		if adaptor.session != nil {
			adaptor.session.Close()
			t.Log("session is not null, close it.")
		}
	}()
	if err := adaptor.Initializer(nil); err != nil {
		t.Errorf("Initializer failed, err: %s\n", err.Error())
		t.Errorf("Initializer failed, can't connect to iotdb with: [host=%s,port=%s]\n", testConfig.ClientConfig.Host, testConfig.ClientConfig.Port)
	}
	if err := adaptor.EventWriter(testEvent); err != nil {
		t.Errorf("Can't write event to iotDB, err: %s\n", err.Error())
	}
}
