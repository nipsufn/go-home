package jobs

import (
	"go-home/cmd/bulb"
	"go-home/config"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/nathan-osman/go-sunrise"
	log "github.com/sirupsen/logrus"
)

func Sunset(scheduler gocron.Scheduler) error {
	jobName := "builtin_sunset"
	for _, job := range scheduler.Jobs() {
		if job.Name() == jobName {
			log.Warnf("sunset job already scheduled")
			return nil
		}
	}
	_, sunset := sunrise.SunriseSunset(
		config.ConfigSingleton.Location.Lat, config.ConfigSingleton.Location.Lon, time.Now().Year(), time.Now().Month(), time.Now().Day())
	log.Infof("scheduling sunset at %v", sunset)
	_, err := scheduler.NewJob(
		gocron.OneTimeJob(
			gocron.OneTimeJobStartDateTime(sunset.Add(time.Minute*-30))),
		gocron.NewTask(func() {
			log.Infof("running sunset routine")
			iStart := uint8(25)
			iStop := uint8(254)
			for i := iStart; i <= iStop; i++ {
				err := bulb.TurnBulbOnByName(i, 2700, "", "all")
				if err != nil {
					log.Errorf("Can't turn on bulb(s): %v", err)
				}
				log.Tracef("iterating sunset routine - %v", i)
				time.Sleep(time.Second * 5)
			}
		}),
		gocron.WithName(jobName))
	if err != nil {
		log.Errorf("could not schedule sunset job")
		return nil
	}
	return nil
}
