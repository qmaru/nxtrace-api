package cmd

import (
	"fmt"
	"os"

	"nxtrace-api/utils"

	"github.com/jessevdk/go-flags"
)

type Option struct {
	Version func()      `short:"v" long:"version" description:"Show version"`
	Mqtt    MqttCommand `command:"mqtt"`
	Web     WebCommand  `command:"web"`
}

func Execute() {
	var opts Option

	parser := flags.NewParser(&opts, flags.Default)
	parser.Name = "nxtapi"
	parser.LongDescription = "nxtrace web server or mqtt client"

	if len(os.Args) == 1 {
		parser.WriteHelp(os.Stdout)
		return
	}

	opts.Version = func() {
		fmt.Printf("%s version %s\n", parser.Name, utils.Version)
		os.Exit(0)
	}

	_, err := parser.Parse()
	if err != nil {
		os.Exit(0)
	}
}
