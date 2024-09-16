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

type Configuration struct {
	Bulb struct {
		Map          map[string]net.IP `mapstructure:"bulbs"`
		Master       string            `mapstructure:"masterBulb"`
		DefaultState struct {
			Brightness  uint8  `mapstructure:"brightness"`
			Temperature uint   `mapstructure:"temperature"`
			Color       string `mapstructure:"color"`
			On          bool   `mapstructure:"on"`
		} `mapstructure:"defaultState"`
	} `mapstructure:"bulb"`
	Serve struct {
		Port uint16 `mapstructure:"port"`
	} `mapstructure:"serve"`
	Schedule struct {
		DB struct {
			Path string `mapstructure:"path"`
		} `mapstructure:"db"`
	} `mapstructure:"schedule"`
	Location struct {
		Lat float64 `mapstructure:"lat"`
		Lon float64 `mapstructure:"lon"`
	} `mapstructure:"location"`
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
		log.WithError(err).Panicf("fatal error marshalling config file")
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
