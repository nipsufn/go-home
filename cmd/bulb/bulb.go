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

	"go-home/config"
)

func getBulbStateByIP(bulb net.IP) error {
	var err error
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
	if err == nil {
		log.Debugf(`Read bulb state: %s`, string(result))
	}

	return err
}

func GetBulbStateByName(bulbs ...string) error {
	var err error
	if len(bulbs) == 0 || bulbs[0] == "all" {
		bulbs = maps.Keys(config.ConfigSingleton.Bulb.Map)
	}
	for _, bulb := range bulbs {
		if ip, ok := config.ConfigSingleton.Bulb.Map[bulb]; ok {
			e := getBulbStateByIP(ip)
			err = errors.Join(e, err)
		}
	}

	return err
}

func turnBulbOnByIP(dimming int64, temperature uint, color string, bulb net.IP) error {
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

	payload.Dimming = dimming
	if temperature != 0 && (temperature < 2700 || temperature > 6500) {
		log.Errorf(`Temperature out of 2700-6500 range: %d`, temperature)
		return errors.New(`temperature out of 2700-6500 range`)
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

	return err
}

func TurnBulbOnByName(brightness uint8, temperature uint, color string, bulbs ...string) error {
	var err error
	if len(bulbs) == 0 || bulbs[0] == "all" {
		bulbs = maps.Keys(config.ConfigSingleton.Bulb.Map)
	}
	for _, bulb := range bulbs {
		if ip, ok := config.ConfigSingleton.Bulb.Map[bulb]; ok {
			var e error
			dimming := int64((float64(brightness) / 255) * 100)
			if 0 < dimming && dimming < 10 {
				dimming = 10
			}
			if dimming == 0 {
				e = TurnBulbOffByName(bulb)
			} else {
				e = turnBulbOnByIP(dimming, temperature, color, ip)
				if e == nil && bulb != config.ConfigSingleton.Bulb.Master {
					config.StateSingleton.Set(bulb, config.BulbState{
						Brightness:  brightness,
						Temperature: temperature,
						Color:       color,
						On:          (dimming > 0)})
				} else if e == nil {
					config.StateSingleton.SetBrightness(bulb, brightness)
					config.StateSingleton.SetTemperature(bulb, temperature)
					config.StateSingleton.SetColor(bulb, color)
				}
			}
			err = errors.Join(e, err)
		}
	}

	return err
}

func TurnBulbOnByState() error {
	var err error
	for _, bulbName := range maps.Keys(config.ConfigSingleton.Bulb.Map) {
		config := config.StateSingleton.Get(bulbName)
		var e error
		if config.On {
			e = TurnBulbOnByName(config.Brightness, config.Temperature, config.Color, bulbName)
		}
		err = errors.Join(e, err)
	}
	return err
}

func turnBulbOffByIP(bulb net.IP) error {
	var (
		response *wizzModels.ResponsePayload
		result   []byte
		err      error
	)

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

	return err
}

func TurnBulbOffByName(bulbs ...string) error {
	var err error
	if len(bulbs) == 0 || bulbs[0] == "all" {
		log.Debugf("Select all bulbs to turn off")
		bulbs = maps.Keys(config.ConfigSingleton.Bulb.Map)
	}
	log.Debugf("Turned lightbulb %v off", bulbs)
	for _, bulb := range bulbs {
		if ip, ok := config.ConfigSingleton.Bulb.Map[bulb]; ok {
			e := turnBulbOffByIP(ip)
			if e == nil && bulb != config.ConfigSingleton.Bulb.Master {
				config.StateSingleton.SetOn(bulb, false)
			}
			err = errors.Join(e, err)
		}
	}

	return err
}
