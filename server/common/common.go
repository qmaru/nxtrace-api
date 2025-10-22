package common

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
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
	ENV_DEBUG       = "TRACE_DEBUG"
	ENV_TIMEOUT     = "TRACE_TIMEOUT"
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

var GetDebug = sync.OnceValue(func() bool {
	var debug bool
	if dbgEnv := os.Getenv(ENV_DEBUG); dbgEnv != "" {
		if dbg, err := strconv.ParseBool(dbgEnv); err == nil && dbg {
			debug = true
		}
	}
	return debug
})

var traceTimeout = getTimeout()

func getTimeout() time.Duration {
	timeoutStr := setDefaultEnv(ENV_TIMEOUT, "120")
	timeoutInt, err := strconv.Atoi(timeoutStr)
	if err != nil || timeoutInt <= 0 {
		timeoutInt = 120
	}
	return time.Duration(timeoutInt) * time.Second
}

func setDefaultEnv(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

func (c *Config) NewCoreConfig() Config {
	return Config{
		TraceCore: setDefaultEnv(ENV_CORE, "/usr/bin/nexttrace"),
	}
}

func (c *Config) NewWebConfig() Config {
	return Config{
		ServerHost: setDefaultEnv(ENV_WEBHOST, "127.0.0.1"),
		ServerPort: setDefaultEnv(ENV_WEBPORT, "8080"),
	}
}

func (c *Config) NewMqttConfig() Config {
	return Config{
		ServerHost:   setDefaultEnv(ENV_MQTTHOST, "127.0.0.1"),
		ServerPort:   setDefaultEnv(ENV_MQTTPORT, "1883"),
		MqttUsername: setDefaultEnv(ENV_MQTTUSER, "qmaru"),
		MqttPassword: setDefaultEnv(ENV_MQTTPASS, "123456"),
		MqttTopic:    setDefaultEnv(ENV_MQTTTOPIC, "trace/data"),
		MqttWithTLS:  setDefaultEnv(ENV_MQTTWITHTLS, "false"),
		MqttClientID: setDefaultEnv(ENV_MQTTCLIENT, "trace"),
	}
}

func RunTrace(host string, params []string) (string, error) {
	config := new(Config)
	coreCfg := config.NewCoreConfig()

	params = append(params, host)
	if GetDebug() {
		log.Printf("[Receive] trace command: %s %s\n", coreCfg.TraceCore, strings.Join(params, " "))
	}

	ctx, cancel := context.WithTimeout(context.Background(), traceTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, coreCfg.TraceCore, params...)
	output, err := cmd.CombinedOutput()

	result := strings.TrimSpace(string(output))
	if ctx.Err() == context.DeadlineExceeded {
		if GetDebug() {
			log.Printf("[Receive] trace timeout=%s result=%s\n", traceTimeout, result)
		}
		return result, fmt.Errorf("timeout=%s", traceTimeout)
	}

	if err != nil {
		if GetDebug() {
			log.Printf("[Receive] trace error=%v result=%s\n", err, result)
		}
		return result, fmt.Errorf("reason=%w", err)
	}

	if GetDebug() {
		log.Printf("[Receive] result: %s\n", result)
	}

	return result, nil
}
