package config

import (
	"net"
	"path/filepath"
	"strings"

	"golang.org/x/exp/maps"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

// TODO: get rid of mapstructure (public archive)
type Configuration struct {
	Bulb struct {
		Map          map[string]net.IP `yaml:"bulbs"`
		Master       string            `yaml:"masterBulb"`
		DefaultState struct {
			Brightness  uint8  `yaml:"brightness"`
			Temperature uint   `yaml:"temperature"`
			Color       string `yaml:"color"`
			On          bool   `yaml:"on"`
		} `yaml:"defaultState"`
	} `yaml:"bulb"`
	Serve struct {
		Port uint16 `yaml:"port"`
	} `yaml:"serve"`
	Schedule struct {
		DB struct {
			Path string `yaml:"path"`
		} `yaml:"db"`
	} `yaml:"schedule"`
	Location struct {
		Lat float64 `yaml:"lat"`
		Lon float64 `yaml:"lon"`
	} `yaml:"location"`
}

var ConfigSingleton Configuration
var StateSingleton State

func Load(path string) error {
	viper.SetConfigName(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)))
	viper.SetConfigType("yaml")
	viper.AddConfigPath(filepath.Dir(path))
	err := viper.ReadInConfig()
	if err != nil {
		log.Panicf("fatal error reading config file: %v", err)
		return err
	}
	var cfg Configuration
	err = viper.Unmarshal(&cfg, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			// Function to support net.IP
			mapstructure.StringToIPHookFunc(),
			// Appended by the two default functions
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	))
	if err != nil {
		log.Panicf("fatal error marshalling config file: %v", err)
		return err
	}
	if cfg.Serve.Port < 1024 {
		log.Panicf("config field `serve.port` must > 1024: %d", cfg.Serve.Port)
	}
	ConfigSingleton = cfg
	var absPath string
	if absPath, err = filepath.Abs(path); err != nil {
		log.Infof("Config loaded from %s", path)
	} else {
		log.Infof("Config loaded from %s", absPath)
	}
	StateSingleton.Init()
	for _, bulbName := range maps.Keys(ConfigSingleton.Bulb.Map) {
		StateSingleton.Set(bulbName, BulbState{
			Brightness:  cfg.Bulb.DefaultState.Brightness,
			Temperature: cfg.Bulb.DefaultState.Temperature,
			Color:       cfg.Bulb.DefaultState.Color,
			On:          cfg.Bulb.DefaultState.On})
	}
	return nil
}
