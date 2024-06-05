package mqtt

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	"nxtrace-api/server/common"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

func SetDot(target string, params []any) (bool, []any) {
	var isFake bool
	var dotPara []any

	if strings.HasPrefix(target, "198.18") {
		isFake = true
		dotPara = []any{"--dot-server", "google"}
	}

	for _, para := range params {
		p := para.(string)
		if p == "--dot-server" {
			dotPara = []any{}
		}
	}

	params = append(params, dotPara...)
	return isFake, params
}

func ParseIP(target string) (string, error) {
	ips, err := net.LookupHost(target)
	if err != nil {
		return "", err
	}

	if len(ips) != 0 {
		return ips[0], nil
	}

	return target, nil
}

var OnConnectionUp = func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
	config := new(common.Config)
	mqttCfg := config.NewMqttConfig()

	topic := mqttCfg.MqttTopic
	clientID := mqttCfg.MqttClientID

	log.Printf("Connecting to %s:%s (tls=%s) \n", mqttCfg.ServerHost, mqttCfg.ServerPort, mqttCfg.MqttWithTLS)
	log.Printf("Subscrib info: id=%s topic=%s\n", clientID, topic)

	ctx := context.Background()
	_, err := cm.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{
				Topic: topic,
				QoS:   0,
			},
		},
	})

	if err != nil {
		if ctx.Err() != nil {
			log.Printf("Subcrib error to topic: %v\n", ctx.Err().Error())
		} else {
			log.Printf("Subcrib info: topic=%s qos=%d\n", topic, 0)
		}
	}
}

var OnConnectError = func(err error) {
	log.Printf("Connection error: msg=%s\n", err.Error())
}

var OnPublishReceived = []func(paho.PublishReceived) (bool, error){
	func(pr paho.PublishReceived) (bool, error) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Task error: %s\n", err)
			}
		}()

		payload := pr.Packet.Payload
		clientID := pr.Client.ClientID()
		topic := pr.Packet.Topic

		if len(payload) != 0 {
			ctx := context.Background()

			data, err := decodeMessage(payload)
			if err != nil {
				log.Panic(err)
			}

			region := data["region"].(string)
			if clientID == region {
				newTopic := fmt.Sprintf("%s/result", topic)
				target := data["target"].(string)
				params := data["params"].([]any)
				sourceName := data["source_name"].(string)
				sourceID := data["source_id"].(string)

				sourceIP, err := ParseIP(target)
				if err != nil {
					log.Panic(err)
				}

				isFake, params := SetDot(sourceIP, params)

				paramsArray := make([]string, 0)
				for _, para := range params {
					paramsArray = append(paramsArray, para.(string))
				}

				log.Printf("Receive a task: [%s] %s (fakeip=%t)\n", region, target, isFake)

				output, err := common.RunTrace(target, paramsArray)
				if err != nil {
					log.Panic(err)
				}

				outputJson := map[string]any{
					"result": output,
					"callback": map[string]string{
						"region":      region,
						"target":      target,
						"source_ip":   sourceIP,
						"source_id":   sourceID,
						"source_name": sourceName,
					},
				}
				pubMessage, err := encodeMessage(outputJson)
				if err != nil {
					log.Panic(err)
				}

				_, err = pr.Client.Publish(ctx, &paho.Publish{
					Topic:   newTopic,
					QoS:     0,
					Retain:  false,
					Payload: pubMessage,
				})

				if err != nil {
					if ctx.Err() != nil {
						log.Panic(ctx.Err())
					}
				} else {
					log.Printf("Publish message: [%s] %s\n", region, newTopic)
				}
			}
		} else {
			log.Printf("Payload error: id=%s topic=%s payload=%s\n", clientID, topic, payload)
		}

		return true, nil
	}}

var OnClientError = func(err error) {
	log.Printf("Client error info: %s\n", err.Error())
}

var OnServerDisconnect = func(d *paho.Disconnect) {
	if d.Properties != nil {
		log.Printf("Server requested disconnect: %s\n", d.Properties.ReasonString)
	} else {
		log.Printf("Server requested disconnect: %d\n", d.ReasonCode)
	}
}
