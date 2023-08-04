package bulb

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"

	"github.com/FerdinaKusumah/wizz"
	wizzModels "github.com/FerdinaKusumah/wizz/models"
	"github.com/spf13/cobra"

	"go-home/config"
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
			return GetBulbStateByName(args...)
		},
	}
	return listCmd
}

func GetBulbStateByIP(bulbs ...net.IP) error {
	var err error
	for _, bulb := range bulbs {
		var (
			response *wizzModels.ResponsePayload
			result   []byte
			e        error
		)
		if response, e = wizz.GetState(bulb.String()); e != nil {
			log.Errorf(`Unable to read response: %s`, e)
			err = errors.Join(e, err)
		}
		if result, e = json.Marshal(response); e != nil {
			log.Errorf(`Unable to convert to json string: %s`, e)
			err = errors.Join(e, err)
		}
		log.Infof(string(result))
	}
	return err
}

func GetBulbStateByName(bulbs ...string) error {
	var err error
	if len(bulbs) == 0 || bulbs[0] == "all" {
		GetBulbStateByIP(maps.Values(config.ConfigSingleton.Bulb.Map)...)
	} else {
		for _, bulb := range bulbs {
			if ip, ok := config.ConfigSingleton.Bulb.Map[bulb]; ok {
				e := GetBulbStateByIP(ip)
				err = errors.Join(e, err)
			}
		}
	}
	return err
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

func TurnBulbOnByIP(brightness uint8, temperature uint, color string, bulbs ...net.IP) error {
	var (
		err      error
		response *wizzModels.ResponsePayload
		result   []byte
		payload  wizzModels.ParamPayload = wizzModels.ParamPayload{
			State: true,
			Speed: 50, // must between 0 - 100
		}
		r, g, b int
	)
	payload.Dimming = int64((float64(brightness) / 255) * 100)
	if 0 < payload.Dimming && payload.Dimming < 10 {
		payload.Dimming = 10
	}
	if payload.Dimming == 0 {
		return TurnBulbOffByIP(bulbs...)
	}

	if temperature != 0 && (temperature < 2500 || temperature > 6500) {
		log.Errorf(`Temperature out of 2500-6500 range: %d`, temperature)
		return errors.New(`temperature out of 2500-6500 range`)
	} else if temperature != 0 {
		payload.ColorTemp = float64(temperature)
	}

	if len(color) > 0 {
		if _, err = fmt.Sscanf(color, "#%02x%02x%02x", &r, &g, &b); err != nil {
			log.Errorf(`Color is not in #RRGGBB format: %s`, color)
			return errors.Join(errors.New(`color is not in #RRGGBB format`), err)
		}
		payload.R = float64(r)
		payload.G = float64(g)
		payload.B = float64(b)
	}

	for _, bulb := range bulbs {
		var e error

		if response, e = wizz.SetPilot(bulb.String(), payload); e != nil {
			log.Errorf(`Unable to read response: %s`, e)
			err = errors.Join(e, err)
		}
		if result, e = json.Marshal(response); e != nil {
			log.Errorf(`Unable to convert to json string: %s`, e)
			err = errors.Join(e, err)
		}
		log.Debugf(string(result))
		log.Infof("Turned lightbulb %s on", bulb.String())
	}
	return err
}

func TurnBulbOnByName(brightness uint8, temperature uint, color string, bulbs ...string) error {
	var err error
	if len(bulbs) == 0 || bulbs[0] == "all" {
		TurnBulbOnByIP(brightness, temperature, color, maps.Values(config.ConfigSingleton.Bulb.Map)...)
	} else {
		for _, bulb := range bulbs {
			if ip, ok := config.ConfigSingleton.Bulb.Map[bulb]; ok {
				e := TurnBulbOnByIP(brightness, temperature, color, ip)
				err = errors.Join(e, err)
			}
		}
	}
	return err
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

func TurnBulbOffByIP(bulbs ...net.IP) error {
	var (
		response *wizzModels.ResponsePayload
		result   []byte
		err      error
	)
	for _, bulb := range bulbs {
		var e error
		if response, e = wizz.TurnOffLight(bulb.String()); e != nil {
			log.Errorf(`Unable to read response: %s`, e)
			err = errors.Join(e, err)
		}
		if result, e = json.Marshal(response); e != nil {
			log.Errorf(`Unable to convert to json string: %s`, e)
			err = errors.Join(e, err)
		}
		log.Debugf(string(result))
		log.Infof("Turned lightbulb %s off", bulb.String())
	}
	return err
}

func TurnBulbOffByName(bulbs ...string) error {
	var err error
	if len(bulbs) == 0 || bulbs[0] == "all" {
		TurnBulbOffByIP(maps.Values(config.ConfigSingleton.Bulb.Map)...)
	} else {
		for _, bulb := range bulbs {
			if ip, ok := config.ConfigSingleton.Bulb.Map[bulb]; ok {
				e := TurnBulbOffByIP(ip)
				err = errors.Join(e, err)
			}
		}
	}
	return err
}
