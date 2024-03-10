package jobs

import (
	"go-home/cmd/bulb"
	"go-home/config"
	"time"

	"github.com/FerdinaKusumah/wizz/connection"
	log "github.com/sirupsen/logrus"
)

func PingMasterBulbRoutine() {
	log.Debugf("Master bulb loop starting")
	for {
		connection.TimeoutMs = 500
		err := bulb.GetBulbStateByName(config.ConfigSingleton.Bulb.Master)

		if err != nil && !config.StateSingleton.GetMasterState() {
			log.Tracef("Master bulb already off")
		} else if err != nil && config.StateSingleton.GetMasterState() {
			log.Tracef("Master bulb set off")
			bulb.TurnBulbOffByName("all")
			config.StateSingleton.SetMasterState(false)
		} else if err == nil && !config.StateSingleton.GetMasterState() {
			log.Tracef("Master bulb set on")
			bulb.TurnBulbOnByState()
			config.StateSingleton.SetMasterState(true)
		} else {
			log.Tracef("Master bulb already on")
		}
		time.Sleep(time.Millisecond * 1500)
	}
}
