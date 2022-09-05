package adapter

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/apache/iotdb-client-go/client"
	"github.com/apache/iotdb-client-go/rpc"
	"github.com/edgexfoundry/edgexfoundry-iotdb-connector/internal/container"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
)

const (
	defaultEdgeXStorageGroup = "root.edgexfoundry"
)

var EdgeXAdapterName = di.TypeInstanceToName(EdgeXFoundryAdapter{})

func EdgeXAdapterFrom(get di.Get) *EdgeXFoundryAdapter {
	return get(EdgeXAdapterName).(*EdgeXFoundryAdapter)
}

var defaultConfig = AdaptorConfig{
	StorageGroup: defaultEdgeXStorageGroup,
	ClientConfig: client.Config{
		Host:     "localhost",
		Port:     "6667",
		UserName: "root",
		Password: "root",
	},
}

type EdgeXFoundryAdapter struct {
	session *client.Session
	Config  AdaptorConfig
}

type AdaptorConfig struct {
	StorageGroup string
	ClientConfig client.Config
	Timeout      time.Duration
}

func NewAdaptor() *EdgeXFoundryAdapter {
	return &EdgeXFoundryAdapter{}
}

func NewAdaptorWithConfig(config AdaptorConfig) *EdgeXFoundryAdapter {
	return &EdgeXFoundryAdapter{
		Config: config,
	}
}

func (adapter *EdgeXFoundryAdapter) BootstrapHandler(ctx context.Context, wg *sync.WaitGroup, startupTimer startup.Timer, dic *di.Container) bool {

	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	err := adapter.Initializer(dic)
	if err != nil {
		lc.Errorf("Failed to connect to iotdb, %v", err)
		return false
	}

	return true
}

func (adapter *EdgeXFoundryAdapter) Initializer(dic *di.Container) error {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)
	clientConfig := &client.Config{
		Host:     config.Databases["Primary"].Host,
		Port:     strconv.Itoa(config.Databases["Primary"].Port),
		UserName: config.Writable.InsecureSecrets["DB"].Secrets["username"],
		Password: config.Writable.InsecureSecrets["DB"].Secrets["password"],
	}
	session := client.NewSession(clientConfig)
	adapter.session = &session
	if err := adapter.session.Open(false, 0); err != nil {
		lc.Errorf("Failed to connect to iotdb, %v", err)
		return err
	}
	if _, err := adapter.session.SetStorageGroup(defaultEdgeXStorageGroup); err != nil {
		lc.Errorf("Failed to set storagegroup to iotdb, %v", err)
		return err
	}
	dic.Update(di.ServiceConstructorMap{
		EdgeXAdapterName: func(get di.Get) interface{} {
			return adapter
		},
	})
	return nil
}

func (adapter *EdgeXFoundryAdapter) EventWriter(event dtos.Event) error {
	cap := len(event.Readings)
	measurements := make([]string, cap)
	dataTypes := make([]client.TSDataType, cap)
	values := make([]interface{}, cap)
	var tsDataType client.TSDataType
	var err error
	for i, reading := range event.Readings {
		measurements[i] = reading.ResourceName
		if tsDataType, err = valueTypeParser(reading.ValueType); err != nil {
			return err
		}
		dataTypes[i] = tsDataType
		values[i], _ = valueConcreteParser(reading.Value, reading.ValueType)
	}

	// deviceName := strings.Split(event.DeviceName, "-")
	// var deviceNameStr string
	// for _, name := range deviceName {
	// 	deviceNameStr += name
	// }

	// fmt.Println(deviceNameStr)
	// deviceIdPath := fmt.Sprintf("%s.%s", defaultEdgeXStorageGroup, deviceNameStr)
	// fmt.Println("deviceIdPath: " + deviceIdPath)

	deviceIdPath := fmt.Sprintf("%s.%s", defaultEdgeXStorageGroup, event.DeviceName)
	status, err := adapter.session.InsertRecord(deviceIdPath, measurements, dataTypes, values, event.Origin)
	if err := checkError(status, err); err != nil {
		return err
	}

	return nil
}

func (adapter *EdgeXFoundryAdapter) CreateMultiTimeseries(event dtos.Event) error {
	cap := len(event.Readings)
	measurements := make([]string, cap)
	dataTypes := make([]client.TSDataType, cap)
	timeseriesPaths := make([]string, cap)
	encodings := make([]client.TSEncoding, cap)
	compressors := make([]client.TSCompressionType, cap)

	var tsDataType client.TSDataType
	var err error
	for i, reading := range event.Readings {
		if tsDataType, err = valueTypeParser(reading.ValueType); err != nil {
			return err
		}
		dataTypes[i] = tsDataType
		measurements[i] = reading.ResourceName
		timeseriesPaths[i] = fmt.Sprintf("%s.%s.%s", defaultEdgeXStorageGroup, event.DeviceName, reading.ResourceName)
		encodings[i] = client.PLAIN
		compressors[i] = client.SNAPPY
	}

	status, err := adapter.session.CreateMultiTimeseries(timeseriesPaths, dataTypes, encodings, compressors)
	if err := checkError(status, err); err != nil {
		return err
	}
	return nil
}

func (adapter *EdgeXFoundryAdapter) ReadingWriter(reading dtos.BaseReading) error {
	return adapter.write(reading)
}

func (adapter *EdgeXFoundryAdapter) write(reading dtos.BaseReading) (err error) {
	measurements := make([]string, 1)
	dataTypes := make([]client.TSDataType, 1)
	values := make([]interface{}, 1)
	var tsDataType client.TSDataType

	measurements[0] = reading.ResourceName
	if tsDataType, err = valueTypeParser(reading.ValueType); err != nil {
		return err
	}
	dataTypes[0] = tsDataType
	values[0] = reading.Value
	if _, err := adapter.session.InsertRecord(reading.DeviceName, measurements, dataTypes, values, reading.Origin); err != nil {
		return err
	}
	return nil
}

func (adapter *EdgeXFoundryAdapter) TsFileSync() error {
	if _, err := exec.LookPath(""); err != nil {
		fmt.Printf("")
	}
	return nil
}

func sortedByTimestamp() {

}

func checkError(status *rpc.TSStatus, err error) error {
	if err != nil {
		return err
	}
	if status != nil {
		if err = client.VerifySuccess(status); err != nil {
			fmt.Println(err.Error())
			return nil
		}
	}
	return nil
}

func valueConcreteParser(value string, valueType string) (interface{}, error) {
	switch valueType {
	case "Bool":
		return strconv.ParseBool(value)
	case "Int32":
		// return strconv.ParseInt(value, 10, 32)
		v, err := strconv.ParseInt(value, 10, 64)
		return int32(v), err
	case "Int64":
		return strconv.ParseInt(value, 10, 64)
	case "Float32":
		return strconv.ParseFloat(value, 32)
	case "Float64":
		return strconv.ParseFloat(value, 64)
	case "String":
		return value, nil
	default:
		return "unknow", errors.New("unsupported value type")
	}
}

func valueTypeParser(valueType string) (client.TSDataType, error) {
	switch valueType {
	case "Bool":
		return client.BOOLEAN, nil
	case "Int32":
		return client.INT32, nil
	case "Int64":
		return client.INT64, nil
	case "Float32":
		return client.FLOAT, nil
	case "Float64":
		return client.DOUBLE, nil
	case "String":
		return client.TEXT, nil
	default:
		return client.UNKNOW, errors.New("unsupported value type")
	}
}
