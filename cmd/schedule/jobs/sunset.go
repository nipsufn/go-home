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
	scheduler.RemoveByTags(jobName)
	// TODO: implement RemoveByName in gocron
	// or commit to using tags
	// for _, job := range scheduler.Jobs() {
	// 	if job.Name() == jobName {
	// 		err := scheduler.RemoveJob(job.ID())
	// 		if err != nil {
	// 			log.WithError(err).Warnf("could not clean up old sunset job")
	// 		}
	// 		log.Debug("cleaned up old sunset jobs")
	// 	}
	// }
	_, sunset := sunrise.SunriseSunset(
		config.ConfigSingleton.Location.Lat, config.ConfigSingleton.Location.Lon, time.Now().Year(), time.Now().Month(), time.Now().Day())
	log.Infof("scheduling sunset job at %v", sunset.Local())
	_, err := scheduler.NewJob(
		gocron.OneTimeJob(
			gocron.OneTimeJobStartDateTime(sunset.Local().Add(time.Minute*-30))),
		gocron.NewTask(func() {
			//check if any bulb is already on
			//if so do nothing, do not override user
			r, e := bulb.GetBulbStateByName()
			if e != nil {
				log.Errorf("%v", e)
				return
			}
			for k, v := range r {
				if v.Result.State {
					log.Infof("bulb %s already on, skipping sunset procedure", k)
					return
				}
			}
			log.Info("running sunset job")
			iStart := uint8(25)
			iStop := uint8(254)
			durationMin := 20
			delaySec := float64(durationMin) * 60 / float64(iStop-iStart)
			for i := iStart; i <= iStop; i++ {
				err := bulb.TurnBulbOnByName(i, 2700, "", "all")
				if err != nil {
					log.Errorf("Can't turn on bulb(s): %v", err)
				}
				log.Tracef("iterating sunset job - iteration %v", i)
				time.Sleep(time.Second * time.Duration(delaySec))
			}
			log.Info("finished sunset job")
		}),
		gocron.WithName(jobName),
		gocron.WithTags(jobName))
	if err != nil {
		log.WithError(err).Error("could not schedule sunset job")
		return nil
	}
	return nil
}
