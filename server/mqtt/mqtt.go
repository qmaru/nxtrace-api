package mqtt

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"time"

	"nxtrace-api/server/common"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func Run() error {
	config := new(common.Config)
	mqttCfg := config.NewMqttConfig()

	if os.Getenv("DEBUG") == "true" {
		// mqtt.DEBUG = log.New(os.Stdout, "", 0)
		mqtt.ERROR = log.New(os.Stdout, "", 0)
	}

	opts := mqtt.NewClientOptions()

	protocol := "ws"
	if mqttCfg.MqttWithTLS == "true" {
		protocol = "wss"
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
		}
		opts.SetTLSConfig(tlsConfig)
	}

	brokerAddr := fmt.Sprintf("%s://%s:%s/mqtt", protocol, mqttCfg.ServerHost, mqttCfg.ServerPort)

	log.Printf("Broker server: %s\n", brokerAddr)

	opts.AddBroker(brokerAddr)
	opts.SetClientID(mqttCfg.MqttClientID)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(5 * time.Second)
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(30 * time.Second)
	opts.SetConnectTimeout(15 * time.Second)
	opts.SetConnectRetry(true)
	opts.SetOnConnectHandler(TraceOnConnectCallback)
	opts.SetReconnectingHandler(TraceReconnectingCallback)
	opts.SetConnectionLostHandler(TraceConnectionLostCallback)

	if mqttCfg.MqttUsername != "" {
		opts.SetUsername(mqttCfg.MqttUsername)
		opts.SetPassword(mqttCfg.MqttPassword)
	}

	c := mqtt.NewClient(opts)

	token := c.Connect()
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}

	n := make(chan struct{})
	<-n

	return nil
}
