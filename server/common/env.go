package common

import (
	"os"
	"strconv"
	"strings"
)

type ServerType byte

const (
	WsType ServerType = iota
	TcpType
)

const (
	ENV_DEBUG           = "TRACE_DEBUG"
	ENV_TIMEOUT         = "TRACE_TIMEOUT"
	ENV_CORE            = "TRACE_CORE"
	ENV_WEB_HOST        = "TRACE_WEB_HOST"
	ENV_WEB_PORT        = "TRACE_WEB_PORT"
	ENV_MQTT_TYPE       = "TRACE_MQTT_TYPE"
	ENV_MQTT_HOST       = "TRACE_MQTT_HOST"
	ENV_MQTT_PORT       = "TRACE_MQTT_PORT"
	ENV_MQTT_USER       = "TRACE_MQTT_USERNAME"
	ENV_MQTT_PASS       = "TRACE_MQTT_PASSWORD"
	ENV_MQTT_TOPIC      = "TRACE_MQTT_TOPIC"
	ENV_MQTT_QOS        = "TRACE_MQTT_QOS"
	ENV_MQTT_RETAIN     = "TRACE_MQTT_RETAIN"
	ENV_MQTT_CLEANSTART = "TRACE_MQTT_CLEANSTART"
	ENV_MQTT_CLIENT     = "TRACE_MQTT_CLIENT"
	ENV_MQTT_WITHTLS    = "TRACE_MQTT_WITHTLS"
)

func (st ServerType) String() string {
	switch st {
	case WsType:
		return "ws"
	case TcpType:
		return "tcp"
	default:
		return "unknown"
	}
}

func getTimeout() int {
	timeoutInt := getEnvInt(ENV_TIMEOUT, 120)
	return timeoutInt
}

func getEnvString(key string, defaultValue string) string {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		return v
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	valStr, exists := os.LookupEnv(key)
	if !exists || strings.TrimSpace(valStr) == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultValue
	}
	return i
}

func getEnvUint8(key string, defaultValue uint8) uint8 {
	valStr, exists := os.LookupEnv(key)
	if !exists || strings.TrimSpace(valStr) == "" {
		return defaultValue
	}
	u, err := strconv.ParseUint(valStr, 10, 8)
	if err != nil {
		return defaultValue
	}
	return uint8(u)
}

func getEnvBool(key string, defaultValue bool) bool {
	valStr, exists := os.LookupEnv(key)
	if !exists || strings.TrimSpace(valStr) == "" {
		return defaultValue
	}
	b, err := strconv.ParseBool(valStr)
	if err != nil {
		return defaultValue
	}
	return b
}

func getEnvServerType(key string, defaultValue ServerType) ServerType {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		s := strings.ToLower(strings.TrimSpace(v))
		switch s {
		case "ws", "websocket":
			return WsType
		case "tcp":
			return TcpType
		default:
			return defaultValue
		}
	}
	return defaultValue
}
