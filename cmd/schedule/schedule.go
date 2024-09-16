package schedule

import (
	"fmt"
	"go-home/config"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	myJobs "go-home/cmd/schedule/jobs"

	"github.com/go-co-op/gocron/v2"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Job struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Name      string         `gorm:"primaryKey"`
	Schedule  string
	OpName    string
}

var (
	db               *gorm.DB
	scheduler        gocron.Scheduler
	signals          chan os.Signal
	onConflictClause = clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"schedule", "op_name"}),
	}
	defaultJobs = []Job{
		{
			Name:     "builtin_sunset_gen",
			Schedule: "00 00 15 * * *",
			OpName:   "sunset",
		},
		{
			Name:     "userdef_wakeup-alarm",
			Schedule: "00 20 07 * * *",
			OpName:   "wakeup",
		},
	}
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

func interruptHandler() {
	signal := <-signals
	log.Infof("caught signal %v", signal)
	var jobs []Job
	for _, job := range scheduler.Jobs() {
		if !strings.HasPrefix(job.Name(), "builtin_") {
			var schedule, opname string
			for _, tag := range job.Tags() {
				if strings.HasPrefix(tag, "SCHEDULE:") {
					schedule = strings.TrimPrefix(tag, "SCHEDULE:")
				}
				if strings.HasPrefix(tag, "OPNAME:") {
					opname = strings.TrimPrefix(tag, "OPNAME:")
				}
			}
			jobs = append(jobs, Job{Name: job.Name(), Schedule: schedule, OpName: opname})
		}
	}
	tx := db.Clauses(onConflictClause).Create(jobs)
	if tx.Error != nil {
		log.WithError(tx.Error).Fatalf("could not persist jobs")
	}
	log.Infof("jobs have been persisted")
	err := scheduler.Shutdown()
	if err != nil {
		log.WithError(err).Error("could not gracefully stop scheduler")
	}
	os.Exit(0)
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
	err = db.AutoMigrate(&Job{})
	if err != nil {
		return err
	}
	var jobs []Job
	db.Find(&jobs)
	// recreate jobs from DB and assert default jobs not present in DB
	// mint the append parameter order
	for _, job := range append(jobs, defaultJobs...) {
		if !doesJobWithNameExist(job.Name) {
			_, err = scheduler.NewJob(
				gocron.CronJob(job.Schedule, true),
				gocron.NewTask(func() { bootstrapJob(job.OpName) }),
				gocron.WithName(job.Name),
				gocron.WithTags("SCHEDULE:"+job.Schedule, "OPNAME:"+job.OpName),
			)
			if err != nil {
				log.WithError(err).Errorf("cannot schedule job `%s`", job.Name)
				return err
			}
			log.Debugf("scheduled job `%s`", job.Name)
		}
	}
	log.Infof("jobs loaded")

	scheduler.Start()
	//go myJobs.PingMasterBulbRoutine()
	log.Infof("scheduler started")

	signals = make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go interruptHandler()
	return nil
}
