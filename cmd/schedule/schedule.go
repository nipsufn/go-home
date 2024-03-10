// TODO:
//   - check: if no sqlite db -> bootstrap, otherwise leave (user might have disabled)
//   - set up default alarm: fake sunrise plus radio (online audio stream) playback (config/sqlite)
package schedule

import (
	"fmt"
	"go-home/cmd/bulb"
	"go-home/config"
	"strconv"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/nathan-osman/go-sunrise"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Job struct {
	gorm.Model
	Id       string
	Schedule string
	OpName   string
}

var (
	db        *gorm.DB
	scheduler gocron.Scheduler
)

func bootstrapJob(opName string) error {
	log.Debugf("in bootstrapJob")
	switch opName {
	case "sunset":
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
	case "wakeup":
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
	default:
		return fmt.Errorf("operation %s is not defined", opName)
	}
}

func doesJobWithNameExist(name string) bool {
	for _, job := range scheduler.Jobs() {
		if job.Name() == name {
			return true
		}
	}
	return false
}

type gocronLogger struct {
}

func (l gocronLogger) Debug(msg string, args ...any) {
	log.Debugf("gocron: %s: %v", msg, args)
}
func (l gocronLogger) Info(msg string, args ...any) {
	log.Infof("gocron: %s: %v", msg, args)
}
func (l gocronLogger) Warn(msg string, args ...any) {
	log.Warnf("gocron: %s: %v", msg, args)
}
func (l gocronLogger) Error(msg string, args ...any) {
	log.Errorf("gocron: %s: %v", msg, args)
}

func StartSchedules() error {
	// TODO: Timezone in config, preferably UNIX style (Europe/Berlin)
	log.Infof("Starting scheduler")
	var err error
	scheduler, err = gocron.NewScheduler(gocron.WithLocation(time.Local), gocron.WithLogger(gocronLogger{}))
	if err != nil {
		log.Fatalf("cannot start scheduler")
	}
	// load jobs from DB
	db, err = gorm.Open(sqlite.Open(config.ConfigSingleton.Schedule.DB.Path), &gorm.Config{})
	if err != nil {
		log.Errorf("cannot open db: %v", err)
		return err
	}
	var jobs []Job
	db.Find(&jobs)
	for _, job := range jobs {
		_, err = scheduler.NewJob(gocron.CronJob(job.Schedule, false), gocron.NewTask(func() { bootstrapJob(job.OpName) }))
		if err != nil {
			log.Errorf("cannot recreate job: %v", err)
			return err
		}
		log.Infof("job %+v", job)
	}
	// check if sunset job exists, create if not
	if !doesJobWithNameExist("builtin_sunset_gen") {
		_, err = scheduler.NewJob(gocron.CronJob("00 00 15 * * *", true), gocron.NewTask(func() { bootstrapJob("sunset") }))
		if err != nil {
			log.Errorf("cannot schedule sunset job: %v", err)
			return err
		}
		log.Infof("scheduled sunset job")
	}

	if !doesJobWithNameExist("userdef_wakeup-alarm") {
		_, err = scheduler.NewJob(gocron.CronJob("00 12 * * *", false), gocron.NewTask(func() { bootstrapJob("wakeup") }))
		if err != nil {
			log.Errorf("cannot schedule wakeup job: %v", err)
			return err
		}
		log.Infof("scheduled wakeup job")
	}
	scheduler.Start()
	// TODO: persist created jobs to db
	//go myJobs.PingMasterBulbRoutine()
	log.Infof("Scheduler started")

	return nil
}
