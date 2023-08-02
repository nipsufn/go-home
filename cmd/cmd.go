package cmd

import (
	"fmt"
	"os"

	"go-home/cmd/bulb"
	"go-home/cmd/serve"
	"go-home/config"

	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

func NewRootCmd() (cmd *cobra.Command) {
	var cfgPath string
	var logLevel string
	cmd = &cobra.Command{
		Use: "go-home",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			level, errB := log.ParseLevel(logLevel)
			if errB != nil {
				log.Fatalf(`Unable to set log level: %s`, errB)
				return errB
			}
			log.SetLevel(level)
			if errA := config.Load(cfgPath); errA != nil {
				log.Fatalf(`Unable to load config: %s`, errA)
				return errA
			}
			return nil
		},
	}
	cmd.PersistentFlags().StringVar(&cfgPath, "config", "./resources/config.yaml", "Path to config file")
	cmd.PersistentFlags().StringVarP(&logLevel, "loglevel", "l", "info", "Log level [panic|fatal|error|warn|warning|info|debug|trace]")
	cmd.AddCommand(bulb.NewBulbCmd())
	cmd.AddCommand(serve.NewServeCmd())

	return cmd
}

func Execute() {
	cmd := NewRootCmd()

	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), err)
		os.Exit(1)
	}
}
