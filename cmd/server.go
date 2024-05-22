package cmd

import (
	"log"

	"nxtrace-server/server/mqtt"
	"nxtrace-server/server/web"

	"github.com/spf13/cobra"
)

var (
	WebCmd = &cobra.Command{
		Use:   "web",
		Short: "Run web server",
		Run: func(cmd *cobra.Command, args []string) {
			err := web.Run()
			if err != nil {
				log.Fatal(err)
			}
		},
	}
	MqttCmd = &cobra.Command{
		Use:   "mqtt",
		Short: "Run mqtt server",
		Run: func(cmd *cobra.Command, args []string) {
			err := mqtt.Run()
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)
