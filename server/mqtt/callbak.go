package mqtt

import (
	"fmt"
	"log"
	"net"
	"strings"

	"nxtrace-api/server/common"

	mqtt "github.com/eclipse/paho.mqtt.golang"
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

var TraceConnectCallback mqtt.OnConnectHandler = func(client mqtt.Client) {
	config := new(common.Config)
	mqttCfg := config.NewMqttConfig()

	log.Printf("Connecting to %s:%s (tls=%s) \n", mqttCfg.ServerHost, mqttCfg.ServerPort, mqttCfg.MqttWithTLS)

	token := client.Subscribe(mqttCfg.MqttTopic, 0, TraceInfoCallback)
	if token.Wait() && token.Error() != nil {
		log.Printf("Subcrib error to topic: %v\n", token.Error())
	} else {
		log.Printf("Subcrib info: id=%s topic=%s qos=%d\n", mqttCfg.MqttClientID, mqttCfg.MqttTopic, 0)
	}
}

var TraceInfoCallback mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Trace info error: %s\n", err)
		}
	}()

	payload := msg.Payload()
	opts := client.OptionsReader()
	clientID := opts.ClientID()
	topic := msg.Topic()

	if len(payload) != 0 {
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

			log.Printf("receive a task: [%s] %s (fakeip=%t)\n", region, target, isFake)

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

			pubToken := client.Publish(newTopic, 0, false, pubMessage)
			if pubToken.Wait() && pubToken.Error() != nil {
				log.Panic(pubToken.Error())
			} else {
				log.Printf("publish message: [%s] %s\n", region, newTopic)
			}
		}
	} else {
		log.Printf("payload error: id=%s topic=%s payload=%s\n", clientID, topic, payload)
	}
}
