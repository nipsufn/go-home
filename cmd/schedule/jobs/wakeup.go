package jobs

import (
	"go-home/cmd/bulb"
	"math"
	"time"

	"github.com/go-co-op/gocron/v2"
	log "github.com/sirupsen/logrus"
)

func Wakeup(scheduler gocron.Scheduler) error {
	log.Infof("running wakeup routine")
	var i uint8
	// range from 2700 to 6500
	kStart := uint(2700)
	kStop := uint(6500)
	// k = a*i + b
	iStart := uint8(25)
	iStop := uint8(254)
	a := float64(kStop-kStart) / float64(iStop-iStart)
	b := float64(iStart) * -a
	durationMin := 20
	delayNs := float64(durationMin) * 60 / float64(iStop-iStart) * math.Pow10(9)
	for i = iStart; i <= iStop; i++ {
		err := bulb.TurnBulbOnByName(i, kStart+uint(a*float64(i)+b), "", "all")
		if err != nil {
			log.Errorf("Can't turn on bulb(s): %v", err)
		}
		log.Tracef("iterating wakeup routine - %v", i)
		time.Sleep(time.Nanosecond * time.Duration(delayNs))
	}

	log.Infof("finished wakeup routine")
	return nil
}
