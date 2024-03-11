package jobs

import (
	"go-home/config"
	"strconv"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/nathan-osman/go-sunrise"
	log "github.com/sirupsen/logrus"
)

func Sunset(scheduler gocron.Scheduler) error {
	log.Infof("scheduling sunset")
	// TODO: move this validation to config load
	lat, err := strconv.ParseFloat(config.ConfigSingleton.Location.Lat, 64)
	if err != nil {
		log.Errorf("could not convert latitude %s to float", config.ConfigSingleton.Location.Lat)
		return nil
	}
	lon, err := strconv.ParseFloat(config.ConfigSingleton.Location.Lon, 64)
	if err != nil {
		log.Errorf("could not convert longitude %s to float", config.ConfigSingleton.Location.Lon)
		return nil
	}
	_, sunset := sunrise.SunriseSunset(
		lat, lon, time.Now().Year(), time.Now().Month(), time.Now().Day())
	_, err = scheduler.NewJob(
		gocron.OneTimeJob(
			gocron.OneTimeJobStartDateTime(sunset.Add(time.Minute*-30))),
		gocron.NewTask(func() {
			log.Infof("running sunset routine")
		}))
	if err != nil {
		log.Errorf("could not schedule sunset job")
		return nil
	}
	return nil
}
