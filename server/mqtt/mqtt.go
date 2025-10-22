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
	config := new(common.Config)
	mqttCfg := config.NewMqttConfig()

	opts := new(autopaho.ClientConfig)

	if dbgEnv := os.Getenv("DEBUG"); dbgEnv != "" {
		if dbg, err := strconv.ParseBool(dbgEnv); err == nil && dbg {
			// opts.Debug = log.New(os.Stdout, "", 0)
			opts.Errors = log.New(os.Stdout, "", 0)
		}
	}

	protocol := "ws"
	if tlsEnv := mqttCfg.MqttWithTLS; tlsEnv != "" {
		if tlsOn, err := strconv.ParseBool(tlsEnv); err == nil && tlsOn {
			protocol = "wss"
			tlsConfig := &tls.Config{
				InsecureSkipVerify: false,
			}
			opts.TlsCfg = tlsConfig
		}
	}

	brokerAddr := &url.URL{
		Scheme: protocol,
		Host:   net.JoinHostPort(mqttCfg.ServerHost, mqttCfg.ServerPort),
		Path:   "/mqtt",
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
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Printf("shutdown signal received, cancelling context")
		cancel()
	}()

	<-ctx.Done()
	log.Printf("context canceled, exiting")

	return nil
}
