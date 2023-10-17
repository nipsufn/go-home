package config

import (
	"net"
	"path/filepath"
	"strings"
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

type MasterBulbState struct {
	mu sync.RWMutex
	on bool
}

func (d *MasterBulbState) Set(isOn bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.on = isOn
}

func (d *MasterBulbState) Get() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.on
}

type Configuration struct {
	Bulb struct {
		Map    map[string]net.IP `mapstructure:"bulbs"`
		Master string            `mapstructure:"masterBulb"`
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
		Lat string `mapstructure:"lat"`
		Lon string `mapstructure:"lon"`
	} `mapstructure:"location"`
}

var ConfigSingleton Configuration

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
	return nil
}
