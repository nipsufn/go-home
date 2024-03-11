package jobs

import (
	"go-home/cmd/bulb"
	"time"

	"github.com/go-co-op/gocron/v2"
	log "github.com/sirupsen/logrus"
)

func Wakeup(scheduler gocron.Scheduler) error {
	log.Infof("running wakeup routine")
	var i uint8
	// range from 0 to ~3800
	// k = a*i + b
	iStart := uint8(25)
	iStop := uint8(254)
	a := 3800.0 / float64(iStop-iStart)
	b := float64(iStart) * -a
	for i = iStart; i <= iStop; i++ {
		err := bulb.TurnBulbOnByName(i, 2700+uint(a*float64(i)+b), "", "all")
		if err != nil {
			log.Errorf("Can't turn on bulb(s): %v", err)
		}
		time.Sleep(time.Second * 1)
	}
	return nil
}
