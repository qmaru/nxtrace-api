package common

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

type TraceConfig struct {
	Path    string
	Timeout int
}

type WebConfig struct {
	ServerHost string
	ServerPort int
}

type MqttConfig struct {
	ServerType     ServerType
	ServerHost     string
	ServerPort     int
	MqttUsername   string
	MqttPassword   string
	MqttWithTLS    bool
	MqttClientID   string
	MqttTopic      string
	MqttQos        uint8
	MqttRetain     bool
	MqttCleanStart bool
}

type Config struct {
	Debug bool
	Trace TraceConfig
	Web   WebConfig
	Mqtt  MqttConfig
}

var NxtConfig *Config

func init() {
	NxtConfig = NewConfig()
}

func NewConfig() *Config {
	cfg := &Config{}

	cfg.Debug = getEnvBool(ENV_DEBUG, false)

	cfg.Trace = TraceConfig{
		Path:    getEnvString(ENV_CORE, "/usr/bin/nexttrace"),
		Timeout: getTimeout(),
	}

	cfg.Web = WebConfig{
		ServerHost: getEnvString(ENV_WEB_HOST, "127.0.0.1"),
		ServerPort: getEnvInt(ENV_WEB_PORT, 8080),
	}

	cfg.Mqtt = MqttConfig{
		ServerType:     getEnvServerType(ENV_MQTT_TYPE, WsType),
		ServerHost:     getEnvString(ENV_MQTT_HOST, "127.0.0.1"),
		ServerPort:     getEnvInt(ENV_MQTT_PORT, 1883),
		MqttUsername:   getEnvString(ENV_MQTT_USER, "qmaru"),
		MqttPassword:   getEnvString(ENV_MQTT_PASS, "123456"),
		MqttTopic:      getEnvString(ENV_MQTT_TOPIC, "trace/data"),
		MqttQos:        getEnvUint8(ENV_MQTT_QOS, 0),
		MqttRetain:     getEnvBool(ENV_MQTT_RETAIN, false),
		MqttWithTLS:    getEnvBool(ENV_MQTT_WITHTLS, false),
		MqttCleanStart: getEnvBool(ENV_MQTT_CLEANSTART, false),
		MqttClientID:   getEnvString(ENV_MQTT_CLIENT, "trace"),
	}

	return cfg
}

func (c *Config) GetCoreConfig() TraceConfig {
	return c.Trace
}

func (c *Config) GetWebConfig() WebConfig {
	return c.Web
}

func (c *Config) GetMqttConfig() MqttConfig {
	return c.Mqtt
}

func RunTrace(host string, params []string) (string, error) {
	coreCfg := NxtConfig.GetCoreConfig()

	params = append(params, host)
	if NxtConfig.Debug {
		log.Printf("[Receive] trace command: %s %s\n", coreCfg.Path, strings.Join(params, " "))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(coreCfg.Timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, coreCfg.Path, params...)
	output, err := cmd.CombinedOutput()

	result := strings.TrimSpace(string(output))
	if ctx.Err() == context.DeadlineExceeded {
		if NxtConfig.Debug {
			log.Printf("[Receive] trace timeout=%d result=%s\n", coreCfg.Timeout, result)
		}
		return result, fmt.Errorf("timeout=%d", coreCfg.Timeout)
	}
	if err != nil {
		if NxtConfig.Debug {
			log.Printf("[Receive] trace error=%v result=%s\n", err, result)
		}
		return result, fmt.Errorf("reason=%w", err)
	}

	if NxtConfig.Debug {
		log.Printf("[Receive] result: %s\n", result)
	}

	return result, nil
}
