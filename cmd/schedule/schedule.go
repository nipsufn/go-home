// TODO:
//   - check: if no sqlite db -> bootstrap, otherwise leave (user might have disabled)
//   - set up default alarm: fake sunrise plus radio (online audio stream) playback (config/sqlite)
package schedule

import (
	"time"

	"go-home/cmd/schedule/jobs"

	"github.com/deepaksinghvi/cdule/pkg/cdule"
	log "github.com/sirupsen/logrus"
)

func StartSchedules() {
	// TODO: check if bootstrap with alarm/sunset is needed
	log.Infof("Starting scheduler")
	scheduler := cdule.Cdule{}
	scheduler.NewCdule()
	sunsetJob := jobs.SunsetJob{}
	jobData := make(map[string]string)
	// TODO: create this in daily job that gets/calculates sunset
	// TODO: add location in config for sunset get/calc
	cdule.NewJob(&sunsetJob, jobData).Build("* 18 * * *")
}

func MasterBulbLoop() {
	for {
		time.Sleep(500 * time.Millisecond)
		jobs.PingMasterBulb()
		// TODO: if state is changed, reflect on all other bulbs (remember state / check if before sunset when turning on)
	}
}
