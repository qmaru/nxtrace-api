package cmd

import (
	"nxtrace-api/server/mqtt"
)

type MqttCommand struct{}

func (c *MqttCommand) Execute(args []string) error {
	err := mqtt.Run()
	if err != nil {
		return err
	}
	return nil
}
