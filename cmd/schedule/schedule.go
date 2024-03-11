// TODO:
//   - check: if no sqlite db -> bootstrap, otherwise leave (user might have disabled)
//   - set up default alarm: fake sunrise plus radio (online audio stream) playback (config/sqlite)
package schedule

import (
	"fmt"
	"go-home/config"
	"time"

	myJobs "go-home/cmd/schedule/jobs"

	"github.com/go-co-op/gocron/v2"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Job struct {
	gorm.Model
	Name     string
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
		return myJobs.Sunset(scheduler)
	case "wakeup":
		return myJobs.Wakeup(scheduler)
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
		_, err = scheduler.NewJob(gocron.CronJob(job.Schedule, true), gocron.NewTask(func() { bootstrapJob(job.OpName) }))
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
		_, err = scheduler.NewJob(gocron.CronJob("00 00 12 * * *", true), gocron.NewTask(func() { bootstrapJob("wakeup") }))
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
