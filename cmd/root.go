package cmd

import (
	"fmt"
	"os"

	"nxtrace-api/utils"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:     "nxtrace",
		Short:   "nxtrace api server, mqtt client",
		Version: utils.Version,
	}
)

func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(
		WebCmd,
		MqttCmd,
	)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
