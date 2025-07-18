package jobs

import (
	"go-home/cmd/bulb"
	"go-home/cmd/playback"
	"go-home/config"
	"net/url"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/nathan-osman/go-sunrise"
	log "github.com/sirupsen/logrus"
)

func Wakeup(scheduler gocron.Scheduler) error {
	log.Infof("running wakeup routine")
	durationMin := 20
	sunrise, _ := sunrise.SunriseSunset(
		config.ConfigSingleton.Location.Lat, config.ConfigSingleton.Location.Lon, time.Now().Year(), time.Now().Month(), time.Now().Day())
	if time.Now().Before((sunrise)) {
		go fadeInLights(uint8(durationMin))
	}
	go fadeInRadio()

	log.Infof("finished wakeup routine")
	return nil
}

func fadeInLights(durationMin uint8) {
	var i uint8
	// range from 2700 to 6500
	kStart := uint(2700)
	kStop := uint(5000)
	// k = a*i + b
	iStart := uint8(25)
	iStop := uint8(254)
	a := float64(kStop-kStart) / float64(iStop-iStart)
	b := float64(iStart) * -a
	delaySec := float64(durationMin) * 60 / float64(iStop-iStart)
	log.Tracef("delaySec: %v", delaySec)
	for i = iStart; i <= iStop; i++ {
		err := bulb.TurnBulbOnByName(i, kStart+uint(a*float64(i)+b), "", "all")
		if err != nil {
			log.Errorf("Can't turn on bulb(s): %v", err)
		}
		log.Tracef("iterating wakeup routine - iteration %v", i)
		time.Sleep(time.Second * time.Duration(delaySec))
	}
}

func fadeInRadio() {
	station := config.ConfigSingleton.Radio.DefaultStation
	jazz := config.ConfigSingleton.Radio.Stations[station]
	playback.PlayURL(url.URL(*jazz), 15*time.Minute)
}
