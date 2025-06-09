package bulb

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewBulbCmd() (bulbCmd *cobra.Command) {
	bulbCmd = &cobra.Command{
		Use:   "bulb",
		Short: "Manage lightbulb(s)",
	}
	bulbCmd.AddCommand(newListCmd())
	bulbCmd.AddCommand(newTurnOnCmd())
	bulbCmd.AddCommand(newTurnOffCmd())

	return bulbCmd
}

func newListCmd() (listCmd *cobra.Command) {
	listCmd = &cobra.Command{
		Use:   "list",
		Short: "Turn lightbulb(s) on",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, e := GetBulbStateByName(args...)
			log.Infof("%v", r)
			return e
		},
	}
	return listCmd
}

func newTurnOnCmd() (turnOnCmd *cobra.Command) {
	var brightness uint8
	var temperature uint
	var color string
	turnOnCmd = &cobra.Command{
		Use:   "turnOn",
		Short: "Turn lightbulb(s) on",
		RunE: func(cmd *cobra.Command, args []string) error {
			return TurnBulbOnByName(brightness, temperature, color, args...)
		},
	}
	turnOnCmd.Flags().Uint8VarP(&brightness, "brightness", "b", 0, "Bulb brightness, 0-255")
	turnOnCmd.MarkFlagRequired("brightness")
	turnOnCmd.Flags().UintVarP(&temperature, "temperature", "t", 0, "Bulb color temperature, 2500-6500")
	turnOnCmd.Flags().StringVarP(&color, "color", "c", "", "Bulb color RGB, #RRGGBB")
	turnOnCmd.MarkFlagsMutuallyExclusive("temperature", "color")
	return turnOnCmd
}

func newTurnOffCmd() (turnOnCmd *cobra.Command) {
	turnOnCmd = &cobra.Command{
		Use:   "turnOff",
		Short: "Turn lightbulb(s) off",
		RunE: func(cmd *cobra.Command, args []string) error {
			return TurnBulbOffByName(args...)
		},
	}
	return turnOnCmd
}
