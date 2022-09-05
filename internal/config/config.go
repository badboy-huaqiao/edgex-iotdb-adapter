package config

import (
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
)

var defaultConfigFilePath = "res/configuration.toml"

// ConfigurationStruct contains the configuration properties for the device service.
type ConfigurationStruct struct {
	Writable WritableInfo
	// MessageQueue contains information for connecting to MessageBus which provides alternative way to publish event
	MessageQueue bootstrapConfig.MessageBusInfo
	Databases    map[string]bootstrapConfig.Database
	Service      bootstrapConfig.ServiceInfo
}

type WritableInfo struct {
	LogLevel        string
	InsecureSecrets bootstrapConfig.InsecureSecrets
}

func (c *ConfigurationStruct) UpdateFromRaw(rawConfig interface{}) bool {
	return false
}

func (c *ConfigurationStruct) UpdateWritableFromRaw(rawWritableConfig interface{}) bool {
	return false
}

func (c *ConfigurationStruct) EmptyWritablePtr() interface{} {
	return nil
}

// GetBootstrap returns the configuration elements required by the bootstrap.
func (c ConfigurationStruct) GetBootstrap() bootstrapConfig.BootstrapConfiguration {
	return bootstrapConfig.BootstrapConfiguration{
		Service: c.Service,
	}
}

// GetLogLevel returns the current ConfigurationStruct's log level.
func (c *ConfigurationStruct) GetLogLevel() string {
	return c.Writable.LogLevel
}

// GetRegistryInfo gets the config.RegistryInfo field from the ConfigurationStruct.
func (c *ConfigurationStruct) GetRegistryInfo() bootstrapConfig.RegistryInfo {
	return bootstrapConfig.RegistryInfo{}
}

// GetInsecureSecrets gets the config.InsecureSecrets field from the ConfigurationStruct.
func (c *ConfigurationStruct) GetInsecureSecrets() bootstrapConfig.InsecureSecrets {
	return bootstrapConfig.InsecureSecrets{}
}

// GetTelemetryInfo gets the config.Telemetry section from the ConfigurationStruct
func (c *ConfigurationStruct) GetTelemetryInfo() *bootstrapConfig.TelemetryInfo {
	return &bootstrapConfig.TelemetryInfo{}
}
