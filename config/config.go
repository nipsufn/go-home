package config

import (
	"net"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/exp/maps"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

type MasterBulbState struct {
	mu sync.RWMutex
	on bool
}

func (mbs *MasterBulbState) Set(isOn bool) {
	mbs.mu.Lock()
	defer mbs.mu.Unlock()

	mbs.on = isOn
}

func (mbs *MasterBulbState) Get() bool {
	mbs.mu.RLock()
	defer mbs.mu.RUnlock()

	return mbs.on
}

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
		Lat string `mapstructure:"lat"`
		Lon string `mapstructure:"lon"`
	} `mapstructure:"location"`
}

type BulbState struct {
	On          bool
	Brightness  uint8
	Temperature uint
	Color       string
}

type State struct {
	mu    sync.RWMutex
	bulbs map[string]BulbState
}

func (s *State) Set(name string, state BulbState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.bulbs[name] = state
}
func (s *State) SetOn(name string, state BulbState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.bulbs[name] = state
}
func (s *State) SetBrightness(name string, brightness uint8) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tmp := s.bulbs[name]
	tmp.Brightness = brightness

	s.bulbs[name] = tmp
}
func (s *State) SetTemperature(name string, temperature uint) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tmp := s.bulbs[name]
	tmp.Temperature = temperature

	s.bulbs[name] = tmp
}
func (s *State) SetColor(name string, color string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tmp := s.bulbs[name]
	tmp.Color = color

	s.bulbs[name] = tmp
}

func (s *State) Get(name string) BulbState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.bulbs[name]
}

func (s *State) Init() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	s.bulbs = make(map[string]BulbState)
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
