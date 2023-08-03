package bulb

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"

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
			return getBulbsStateByName(args...)
		},
	}
	return listCmd
}

func getBulbStateByIP(bulb net.IP) error {
	var (
		response *wizzModels.ResponsePayload
		result   []byte
		err      error
	)
	if response, err = wizz.GetState(bulb.String()); err != nil {
		log.Errorf(`Unable to read response: %s`, err)
		return err
	}
	if result, err = json.Marshal(response); err != nil {
		log.Errorf(`Unable to convert to json string: %s`, err)
		return err
	}
	log.Infof(string(result))
	return nil
}

func getBulbsStateByName(bulbs ...string) error {
	var err error
	for key, element := range config.ConfigSingleton.Bulb.List {
		if slices.Contains(bulbs, key) || slices.Contains(bulbs, "all") || len(bulbs) == 0 {
			log.Infof("Checking bulb \"%s\"", key)
			e := getBulbStateByIP(net.ParseIP(element))
			err = errors.Join(e, err)
		}
	}
	return err
}

//func getBulbsStateByIPs(bulbs ...net.IP) error {
//	var err, e error
//	for _, bulb := range bulbs {
//		if e = getBulbStateByIP(bulb); err != nil {
//			err = errors.Join(err, e)
//			continue
//		}
//	}
//	if err != nil {
//		log.Errorf("Unable to get all bulb states.\n %s", err)
//		return err
//	}
//	log.Infof("Got lightbulb state")
//	return err
//}

func newTurnOnCmd() (turnOnCmd *cobra.Command) {
	var brightness uint8
	var temperature uint
	var color string
	turnOnCmd = &cobra.Command{
		Use:   "turnOn",
		Short: "Turn lightbulb(s) on",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			for key, element := range config.ConfigSingleton.Bulb.List {
				if slices.Contains(args, key) || len(args) == 0 {
					log.Infof("Checking bulb \"%s\"", key)
					e := turnOn(net.ParseIP(element), brightness, temperature, color)
					err = errors.Join(e, err)
				}
			}
			return err
		},
	}
	turnOnCmd.Flags().Uint8VarP(&brightness, "brightness", "b", 0, "Bulb brightness, 0-255")
	turnOnCmd.MarkFlagRequired("brightness")
	turnOnCmd.Flags().UintVarP(&temperature, "temperature", "t", 0, "Bulb color temperature, 2500-6500")
	turnOnCmd.Flags().StringVarP(&color, "color", "c", "", "Bulb color RGB, #RRGGBB")
	turnOnCmd.MarkFlagsMutuallyExclusive("temperature", "color")
	return turnOnCmd
}

func turnOn(bulb net.IP, brightness uint8, temperature uint, color string) error {
	var (
		response *wizzModels.ResponsePayload
		payload  wizzModels.ParamPayload = wizzModels.ParamPayload{
			State: true,
			Speed: 50, // must between 0 - 100
		}
		result  []byte
		err     error = nil
		r, g, b int
	)

	payload.Dimming = int64((float64(brightness) / 255) * 100)
	if 0 < payload.Dimming && payload.Dimming < 10 {
		payload.Dimming = 10
	}
	if payload.Dimming == 0 {
		return turnOff(bulb)
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

	if response, err = wizz.SetPilot(bulb.String(), payload); err != nil {
		log.Errorf(`Unable to read response: %s`, err)
		return err
	}
	if result, err = json.Marshal(response); err != nil {
		log.Errorf(`Unable to convert to json string: %s`, err)
		return err
	}
	log.Debugf(string(result))
	log.Infof("Turned lightbulb(s) on")
	return nil
}

func newTurnOffCmd() (turnOnCmd *cobra.Command) {
	turnOnCmd = &cobra.Command{
		Use:   "turnOff",
		Short: "Turn lightbulb(s) off",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			for key, element := range config.ConfigSingleton.Bulb.List {
				if slices.Contains(args, key) || len(args) == 0 {
					log.Infof("Checking bulb \"%s\"", key)
					e := turnOff(net.ParseIP(element))
					err = errors.Join(e, err)
				}
			}
			return err
		},
	}
	return turnOnCmd
}

func turnOff(bulb net.IP) error {
	var response *wizzModels.ResponsePayload
	var result []byte
	var err error = nil
	if response, err = wizz.TurnOffLight(bulb.String()); err != nil {
		log.Errorf(`Unable to read response: %s`, err)
		return err
	}
	if result, err = json.Marshal(response); err != nil {
		log.Errorf(`Unable to convert to json string: %s`, err)
		return err
	}
	log.Debugf(string(result))
	log.Infof("Turned lightbulb(s) off")
	return err
}
