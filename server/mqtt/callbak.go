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
	if strings.HasPrefix(target, "198.18.") {
		isFake = true
		if !slices.Contains(params, "--dot-server") {
			params = append(params, "--dot-server", "google")
		}
	}
	return isFake, params
}

func ParseIP(target string) (string, error) {
	if ip := net.ParseIP(target); ip != nil {
		return ip.String(), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, target)
	if err != nil {
		return "", err
	}

	if len(addrs) == 0 {
		return "", fmt.Errorf("no IP records for %s", target)
	}

	for _, a := range addrs {
		if a.IP.To4() != nil {
			return a.IP.String(), nil
		}
	}

	return addrs[0].IP.String(), nil
}

var OnConnectionUp = func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
	mqttCfg := common.NxtConfig.GetMqttConfig()

	topic := mqttCfg.MqttTopic
	qos := mqttCfg.MqttQos
	retain := mqttCfg.MqttRetain
	clientID := mqttCfg.MqttClientID

	traceTopic := fmt.Sprintf("%s/%s", topic, clientID)
	log.Printf("[UP] connecting to %s:%d (tls=%t)\n", mqttCfg.ServerHost, mqttCfg.ServerPort, mqttCfg.MqttWithTLS)
	log.Printf("[UP] subscribe info: id=%s qos=%d retain=%t clean=%t\n", clientID, qos, retain, mqttCfg.MqttCleanStart)
	log.Printf("[UP] topic info: base_topic=%s trace_topic=%s\n", topic, traceTopic)

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
			log.Printf("[UP] subscribe error to topic: %v\n", ctx.Err().Error())
		}
	}
}

var OnConnectError = func(err error) {
	log.Printf("[ERROR] connection error: msg=%s\n", err.Error())
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
		requestId := task.RequestId
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

		log.Printf("[Receive] request_id=%s region=%s target=%s (fakeip=%t)\n", requestId, region, target, isFake)
		log.Printf("[Receive] request_id=%s params from %s: %+v\n", requestId, sourceName, params)
		log.Printf("[Receive] request_id=%s trace params %v\n", requestId, latestParams)
		log.Printf("[Receive] request_id=%s receive_topic=%s result_topic=%s\n", requestId, topic, resultTopic)

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
			"request_id": requestId,
			"result":     traceResult,
			"callback": map[string]string{
				"region":      region,
				"target":      target,
				"source_ip":   sourceIP,
				"source_id":   sourceID,
				"source_name": sourceName,
			},
		}

		if common.NxtConfig.Debug {
			log.Printf("[Receive] request_id=%s trace publish message: %s\n", requestId, pubTextMessage)
		}

		pubMessage, err := encodeMessage(pubTextMessage)
		if err != nil {
			log.Printf("[Receive] request_id=%s encode output message error: %v\n", requestId, err)
			return true, nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		log.Printf("[Receive] request_id=%s start publish\n", requestId)
		log.Printf("[Receive] request_id=%s id=%s qos=%d retain=%t clean=%t\n", requestId, clientID, mqttCfg.MqttQos, mqttCfg.MqttRetain, mqttCfg.MqttCleanStart)
		_, err = pr.Client.Publish(ctx, &paho.Publish{
			Topic:   resultTopic,
			QoS:     mqttCfg.MqttQos,
			Retain:  mqttCfg.MqttRetain,
			Payload: pubMessage,
		})

		if err != nil {
			log.Printf("[Receive] request_id=%s publish error: topic=%s err=%v\n", requestId, resultTopic, err)
		} else {
			log.Printf("[Receive] request_id=%s publish message: [%s] %s\n", requestId, region, resultTopic)
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
