package mqtt

import (
	"context"
	"fmt"
	"log"
	"net"
	"slices"
	"strings"
	"time"

	"nxtrace-api/server/common"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

func SetDot(target string, params []string) (bool, []string) {
	var isFake bool
	var dotPara []string

	if strings.HasPrefix(target, "198.18") {
		isFake = true
		dotPara = []string{"--dot-server", "google"}
	}

	if slices.Contains(params, "--dot-server") {
		dotPara = []string{}
	}

	if len(dotPara) > 0 {
		params = append(params, dotPara...)
	}
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

	log.Printf("Connecting to %s:%s (tls=%s)\n", mqttCfg.ServerHost, mqttCfg.ServerPort, mqttCfg.MqttWithTLS)
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
				log.Printf("[Receive] task error: %s\n", err)
			}
		}()

		payload := pr.Packet.Payload
		clientID := pr.Client.ClientID()
		topic := pr.Packet.Topic

		if len(payload) == 0 {
			log.Printf("[Receive] payload empty: id=%s topic=%s\n", clientID, topic)
			return true, nil
		}

		task, err := decodeTaskMessage(payload)
		if err != nil {
			log.Printf("[Receive] decode payload error: %v\n", err)
			log.Printf("[Receive] payload %s\n", string(payload))
			return true, nil
		}

		region := task.Region
		if clientID != region {
			return true, nil
		}

		newTopic := fmt.Sprintf("%s/result", topic)
		target := task.Target
		params := task.Params
		sourceName := task.SourceName
		sourceID := task.SourceID

		sourceIP, err := ParseIP(target)
		if err != nil {
			log.Printf("[Receive] parse ip(%s) error: %v\n", target, err)
			sourceIP = target
			return true, nil
		}

		isFake, latestParams := SetDot(sourceIP, params)

		log.Printf("[Receive] region=%s target=%s (fakeip=%t)\n", region, target, isFake)
		log.Printf("[Receive] params from %s: %+v\n", sourceName, params)
		log.Printf("[Receive] trace params %v\n", latestParams)

		var traceResult string

		output, err := common.RunTrace(target, latestParams)
		if err != nil {
			log.Printf("[Receive] trace error: %v\n", err)
			if output == "" {
				errOutput := map[string]string{
					"error": err.Error(),
				}
				errOutputJson, err := encodeMessage(errOutput)
				if err != nil {
					log.Printf("[Receive] encode error message error: %v\n", err)
					return true, nil
				}
				traceResult = string(errOutputJson)
			}
		} else {
			traceResult = output
		}

		pubTextMessage := map[string]any{
			"result": traceResult,
			"callback": map[string]string{
				"region":      region,
				"target":      target,
				"source_ip":   sourceIP,
				"source_id":   sourceID,
				"source_name": sourceName,
			},
		}

		if common.GetDebug() {
			log.Printf("[Receive] trace publish message: %s\n", pubTextMessage)
		}

		pubMessage, err := encodeMessage(pubTextMessage)
		if err != nil {
			log.Printf("[Receive] encode output message error: %v\n", err)
			return true, nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		log.Printf("[Receive] start publish\n")
		_, err = pr.Client.Publish(ctx, &paho.Publish{
			Topic:   newTopic,
			QoS:     0,
			Retain:  false,
			Payload: pubMessage,
		})

		if err != nil {
			log.Printf("[Receive] publish error: topic=%s err=%v\n", newTopic, err)
		} else {
			log.Printf("[Receive] publish message: [%s] %s\n", region, newTopic)
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
