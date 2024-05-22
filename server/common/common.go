package common

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Config struct {
	TraceCore    string
	ServerHost   string
	ServerPort   string
	MqttUsername string
	MqttPassword string
	MqttTopic    string
	MqttClientID string
	MqttWithTLS  string
}

const (
	ENV_CORE        = "TRACE_CORE"
	ENV_WEBHOST     = "TRACE_WEB_HOST"
	ENV_WEBPORT     = "TRACE_WEB_PORT"
	ENV_MQTTHOST    = "TRACE_MQTT_HOST"
	ENV_MQTTPORT    = "TRACE_MQTT_PORT"
	ENV_MQTTUSER    = "TRACE_MQTT_USERNAME"
	ENV_MQTTPASS    = "TRACE_MQTT_PASSWORD"
	ENV_MQTTTOPIC   = "TRACE_MQTT_TOPIC"
	ENV_MQTTCLIENT  = "TRACE_MQTT_CLIENT"
	ENV_MQTTWITHTLS = "TRACE_MQTT_WITHTLS"
)

func (c *Config) setDefaultEnv(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

func (c *Config) NewCoreConfig() Config {
	return Config{
		TraceCore: c.setDefaultEnv(ENV_CORE, "/usr/bin/nexttrace"),
	}
}

func (c *Config) NewWebConfig() Config {
	return Config{
		ServerHost: c.setDefaultEnv(ENV_WEBHOST, "127.0.0.1"),
		ServerPort: c.setDefaultEnv(ENV_WEBPORT, "8080"),
	}
}

func (c *Config) NewMqttConfig() Config {
	return Config{
		ServerHost:   c.setDefaultEnv(ENV_MQTTHOST, "127.0.0.1"),
		ServerPort:   c.setDefaultEnv(ENV_MQTTPORT, "1883"),
		MqttUsername: c.setDefaultEnv(ENV_MQTTUSER, "qmaru"),
		MqttPassword: c.setDefaultEnv(ENV_MQTTPASS, "123456"),
		MqttTopic:    c.setDefaultEnv(ENV_MQTTTOPIC, "trace/data"),
		MqttWithTLS:  c.setDefaultEnv(ENV_MQTTWITHTLS, "false"),
		MqttClientID: c.setDefaultEnv(ENV_MQTTCLIENT, fmt.Sprintf("qmeta-pub-%d", time.Now().UnixMilli())),
	}
}

func RunTrace(host string, params []string) (string, error) {
	config := new(Config)
	coreCfg := config.NewCoreConfig()

	params = append(params, host)
	cmd := exec.Command(coreCfg.TraceCore, params...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	results := strings.TrimSpace(string(output))
	return results, nil
}
