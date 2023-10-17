// TODO:
//   - check: if no sqlite db -> bootstrap, otherwise leave (user might have disabled)
//   - set up default alarm: fake sunrise plus radio (online audio stream) playback (config/sqlite)
package schedule

import (
	"fmt"
	"go-home/cmd/bulb"
	myJobs "go-home/cmd/schedule/jobs"
	"go-home/config"
	"strconv"
	"time"

	"github.com/go-co-op/gocron"
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
	scheduler *gocron.Scheduler
)

func bootstrapJob(opName string) error {
	switch opName {
	case "sunset":
		log.Infof("scheduling sunset")
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
		scheduler.At(sunset.Add(time.Minute * -30)).Tag("builtin_sunset_job").Do(func() {
			return
		})
		return nil
	case "wakeup":
		log.Infof("running wakeup routine")
		var i uint8
		for i = 25; i <= 255; i++ {
			// range from 0 to ~4000
			tempOffset := 4000 * (((140 - 25) / 255) * 1.108696)
			bulb.TurnBulbOnByName(i, 2500+uint(tempOffset), "")
			time.Sleep(time.Second * 5)
		}
		return nil
	default:
		return fmt.Errorf("operation %s is not defined", opName)
	}
}

func doesJobWithTagExist(tag string) bool {
	jobs, err := scheduler.FindJobsByTag(tag)
	if err != nil || len(jobs) == 0 {
		return false
	}
	return true
}

func StartSchedules(masterBulbState *config.MasterBulbState) error {
	// TODO: Timezone in config, preferably UNIX style (Europe/Berlin)
	log.Infof("Starting scheduler")
	scheduler = gocron.NewScheduler(time.Local)
	scheduler.TagsUnique()
	// load jobs from DB
	var err error
	db, err = gorm.Open(sqlite.Open(config.ConfigSingleton.Schedule.DB.Path), &gorm.Config{})
	if err != nil {
		log.Errorf("cannot open db: %v", err)
		return err
	}
	var jobs []Job
	db.Find(&jobs)
	for _, job := range jobs {
		job, err := scheduler.Cron(job.Schedule).Tag(job.Id).Do(func() { bootstrapJob(job.OpName) })
		if err != nil {
			log.Errorf("cannot recreate job: %v", err)
			return err
		}
		log.Infof("job %+v", job)
	}
	// check if sunset job exists, create if not
	if !doesJobWithTagExist("builtin_sunset_gen") {
		scheduler.Cron("00 00 15 * * *").Tag("builtin_sunset_gen").Do(func() { bootstrapJob("sunset") })
	}

	if !doesJobWithTagExist("userdef_wakeup-alarm") {
		scheduler.Cron("00 30 06 * * *").Tag("userdef_wakeup-alarm").Do(func() { bootstrapJob("wakeup") })
	}
	// TODO: persist created jobs to db
	go myJobs.PingMasterBulbRoutine(masterBulbState)
	log.Infof("Scheduler started")

	return nil
}
