package mqtt

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"nxtrace-api/server/common"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

func Run() error {
	mqttCfg := common.NxtConfig.GetMqttConfig()

	opts := new(autopaho.ClientConfig)

	if common.NxtConfig.Debug {
		// opts.Debug = log.New(os.Stdout, "", 0)
		opts.Errors = log.New(os.Stdout, "", 0)
		log.Println("MQTT client starting in debug mode")
	}

	var protocol string
	var path string
	switch mqttCfg.ServerType {
	case common.WsType:
		protocol = "ws"
		path = "/mqtt"
		if mqttCfg.MqttWithTLS {
			protocol = "wss"
			tlsConfig := &tls.Config{
				InsecureSkipVerify: false,
			}
			opts.TlsCfg = tlsConfig
		}
	case common.TcpType:
		protocol = "mqtt"
		if mqttCfg.MqttWithTLS {
			protocol = "mqtts"
			tlsConfig := &tls.Config{
				InsecureSkipVerify: false,
			}
			opts.TlsCfg = tlsConfig
		}
	default:
		return nil
	}

	brokerAddr := &url.URL{
		Scheme: protocol,
		Host:   net.JoinHostPort(mqttCfg.ServerHost, strconv.Itoa(mqttCfg.ServerPort)),
		Path:   path,
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
	opts.CleanStartOnInitialConnection = mqttCfg.MqttCleanStart

	if mqttCfg.MqttUsername != "" {
		opts.ConnectUsername = mqttCfg.MqttUsername
		opts.ConnectPassword = []byte(mqttCfg.MqttPassword)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := autopaho.NewConnection(ctx, *opts)
	if err != nil {
		return err
	}

	connectCtx, connectCancel := context.WithTimeout(ctx, 60*time.Second)
	defer connectCancel()

	if err = client.AwaitConnection(connectCtx); err != nil {
		return err
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sigCh
		log.Println("shutdown signal received, cancelling context")
		cancel()
	}()

	<-ctx.Done()
	log.Println("context canceled, exiting")

	return nil
}
