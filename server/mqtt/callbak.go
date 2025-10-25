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
	mqttCfg := common.NxtConfig.GetMqttConfig()

	topic := mqttCfg.MqttTopic
	qos := mqttCfg.MqttQos
	retain := mqttCfg.MqttRetain
	clientID := mqttCfg.MqttClientID

	traceTopic := fmt.Sprintf("%s/%s", topic, clientID)
	log.Printf("Connecting to %s:%d (tls=%t)\n", mqttCfg.ServerHost, mqttCfg.ServerPort, mqttCfg.MqttWithTLS)
	log.Printf("Subscribe info: id=%s qos=%d retain=%t\n clean=%t", clientID, qos, retain, mqttCfg.MqttCleanStart)
	log.Printf("Topic info: base_topic=%s trace_topic=%s\n", topic, traceTopic)

	ctx := context.Background()
	_, err := cm.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{
				Topic:             traceTopic,
				QoS:               qos,
				RetainAsPublished: retain,
			},
		},
	})

	if err != nil {
		if ctx.Err() != nil {
			log.Printf("Subscribe error to topic: %v\n", ctx.Err().Error())
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

		mqttCfg := common.NxtConfig.GetMqttConfig()

		resultTopic := fmt.Sprintf("%s/result", mqttCfg.MqttTopic)
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
		log.Printf("[Receive] receive_topic=%s result_topic=%s\n", topic, resultTopic)

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

		if common.NxtConfig.Debug {
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
			Topic:   resultTopic,
			QoS:     mqttCfg.MqttQos,
			Retain:  mqttCfg.MqttRetain,
			Payload: pubMessage,
		})

		if err != nil {
			log.Printf("[Receive] publish error: topic=%s err=%v\n", resultTopic, err)
		} else {
			log.Printf("[Receive] publish message: [%s] %s\n", region, resultTopic)
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
