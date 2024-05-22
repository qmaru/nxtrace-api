package mqtt

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"time"

	"nxtrace-server/server/common"

	"github.com/eclipse/paho.mqtt.golang"
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

	opts.AddBroker(brokerAddr)
	opts.SetClientID(mqttCfg.MqttClientID)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetCleanSession(false)
	opts.ConnectRetry = true
	opts.AutoReconnect = true
	opts.OnConnect = TraceConnectCallback
	if mqttCfg.MqttUsername != "" {
		opts.Username = mqttCfg.MqttUsername
		opts.Password = mqttCfg.MqttPassword
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
