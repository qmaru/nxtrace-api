package cmd

import (
	"nxtrace-api/server/mqtt"
	"nxtrace-api/server/web"
)

type MqttCommand struct{}

func (c *MqttCommand) Execute(args []string) error {
	err := mqtt.Run()
	if err != nil {
		return err
	}
	return nil
}

type WebCommand struct{}

func (c *WebCommand) Execute(args []string) error {
	err := web.Run()
	if err != nil {
		return err
	}
	return nil
}
