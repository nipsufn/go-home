package jobs

import (
	"go-home/config"

	probing "github.com/prometheus-community/pro-bing"
	log "github.com/sirupsen/logrus"
)

func PingMasterBulb() error {
	pinger, err := probing.NewPinger(config.ConfigSingleton.Bulb.List[config.ConfigSingleton.Bulb.Master])
	if err != nil {
		log.Errorf("Unable to create pinger: %s", err)
	}
	pinger.Count = 1
	return pinger.Run()
}
