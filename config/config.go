package config

import (
	"path/filepath"
	"strings"

	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

type Configuration struct {
	Schedule struct {
		Cduletype        string `yaml:"cduletype"`
		Dburl            string `yaml:"dburl"`
		Cduleconsistency string `yaml:"cduleconsistency"`
	} `yaml:"schedule"`
	Bulbs      map[string]string `yaml:"bulbs"`
	MasterBulb string            `yaml:"masterBulb"`
}

var ConfigSingleton Configuration

func Load(path string) error {
	viper.SetConfigName(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)))
	viper.SetConfigType("yaml")
	viper.AddConfigPath(filepath.Dir(path))
	err := viper.ReadInConfig()
	if err != nil {
		log.Panicf("fatal error reading config file: %w", err)
		return err
	}
	var cfg Configuration
	err = viper.Unmarshal(&cfg)
	if err != nil {
		log.Panicf("fatal error marshalling config file: %w", err)
		return err
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
