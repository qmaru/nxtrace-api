package mqtt

import (
	"context"
	"fmt"
	"log"
	"net"
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

	for _, p := range params {
		if p == "--dot-server" {
			dotPara = []string{}
			break
		}
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
				log.Printf("task error: %s\n", err)
			}
		}()

		payload := pr.Packet.Payload
		clientID := pr.Client.ClientID()
		topic := pr.Packet.Topic

		if len(payload) == 0 {
			log.Printf("Payload empty: id=%s topic=%s\n", clientID, topic)
			return true, nil
		}

		task, err := decodeTaskMessage(payload)
		if err != nil {
			log.Printf("decodeTaskMessage error: %v\n", err)
			return true, nil
		}

		region := task.Region
		if clientID != region {
			return true, nil
		}

		newTopic := fmt.Sprintf("%s/result", topic)
		target := task.Target
		paramsArray := task.Params
		sourceName := task.SourceName
		sourceID := task.SourceID

		sourceIP, err := ParseIP(target)
		if err != nil {
			log.Printf("ParseIP(%s) error: %v\n", target, err)
			sourceIP = target
			return true, nil
		}

		isFake, params := SetDot(sourceIP, paramsArray)

		log.Printf("Receive a task: [%s] %s (fakeip=%t)\n", region, target, isFake)
		log.Printf("Receive task params from %s: %+v\n", sourceName, params)

		output, err := common.RunTrace(target, paramsArray)
		if err != nil {
			log.Printf("RunTrace error: %v\n", err)
			output = err.Error()
		}

		log.Printf("RunTrace complete\n")

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

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		log.Printf("Start publish message\n")
		_, err = pr.Client.Publish(ctx, &paho.Publish{
			Topic:   newTopic,
			QoS:     0,
			Retain:  false,
			Payload: pubMessage,
		})

		if err != nil {
			log.Printf("Publish error topic=%s err=%v\n", newTopic, err)
		} else {
			log.Printf("Publish message: [%s] %s\n", region, newTopic)
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
