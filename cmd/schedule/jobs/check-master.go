package jobs

import (
	"go-home/cmd/bulb"
	"go-home/config"
	"time"

	"github.com/FerdinaKusumah/wizz/connection"
	log "github.com/sirupsen/logrus"
)

func PingMasterBulbRoutine(bulbState *config.MasterBulbState) {
	log.Debugf("Master bulb loop starting")
	for {
		connection.TimeoutMs = 500
		err := bulb.GetBulbStateByName(config.ConfigSingleton.Bulb.Master)
		if err != nil && !bulbState.Get() {
			log.Tracef("Master bulb already off")
		} else if err != nil && bulbState.Get() {
			log.Tracef("Master bulbb set off")
			bulb.TurnBulbOffByName("all")
			bulbState.Set(false)
		} else if err == nil && !bulbState.Get() {
			log.Tracef("Master bulb set on")
			bulb.TurnBulbOnByName(255, 5000, "")
			bulbState.Set(true)
		} else {
			log.Tracef("Master bulb already on")
		}
		time.Sleep(time.Millisecond * 1500)
	}
}
