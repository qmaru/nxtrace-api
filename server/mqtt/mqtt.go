package mqtt

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/url"
	"os"

	"nxtrace-api/server/common"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

func Run() error {
	config := new(common.Config)
	mqttCfg := config.NewMqttConfig()

	opts := new(autopaho.ClientConfig)

	if os.Getenv("DEBUG") == "true" {
		// opts.Debug = log.New(os.Stdout, "", 0)
		opts.Errors = log.New(os.Stdout, "", 0)
	}

	protocol := "ws"
	if mqttCfg.MqttWithTLS == "true" {
		protocol = "wss"
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
		}
		opts.TlsCfg = tlsConfig
	}

	brokerAddr, err := url.Parse(fmt.Sprintf("%s://%s:%s/mqtt", protocol, mqttCfg.ServerHost, mqttCfg.ServerPort))
	if err != nil {
		return err
	}

	log.Printf("Broker server: %s\n", brokerAddr)

	opts.ServerUrls = []*url.URL{brokerAddr}
	opts.KeepAlive = 60
	opts.CleanStartOnInitialConnection = true
	opts.SessionExpiryInterval = 0xFFFFFFFF
	opts.OnConnectionUp = OnConnectionUp
	opts.OnConnectError = OnConnectError
	opts.ClientConfig = paho.ClientConfig{
		ClientID:           mqttCfg.MqttClientID,
		OnPublishReceived:  OnPublishReceived,
		OnClientError:      OnClientError,
		OnServerDisconnect: OnServerDisconnect,
	}

	if mqttCfg.MqttUsername != "" {
		opts.ConnectUsername = mqttCfg.MqttUsername
		opts.ConnectPassword = []byte(mqttCfg.MqttPassword)
	}

	ctx := context.Background()
	client, err := autopaho.NewConnection(ctx, *opts)
	if err != nil {
		return err
	}

	if err = client.AwaitConnection(ctx); err != nil {
		return err
	}

	n := make(chan struct{})
	<-n

	return nil
}
