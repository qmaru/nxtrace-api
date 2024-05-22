package mqtt

import (
	"fmt"
	"log"

	"nxtrace-server/server/common"

	"github.com/eclipse/paho.mqtt.golang"
)

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

			log.Printf("receive a task: [%s] %s\n", region, target)

			paramsArray := make([]string, 0)
			for _, para := range params {
				paramsArray = append(paramsArray, para.(string))
			}

			output, err := common.RunTrace(target, paramsArray)
			if err != nil {
				log.Panic(err)
			}

			outputJson := map[string]any{
				"result": output,
				"callback": map[string]string{
					"region":      region,
					"target":      target,
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
