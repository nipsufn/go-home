package jobs

import (
	"strconv"

	"github.com/deepaksinghvi/cdule/pkg/cdule"
	log "github.com/sirupsen/logrus"
)

var sunsetJobData map[string]string

type SunsetJob struct {
	Job cdule.Job
}

func (m SunsetJob) Execute(jobData map[string]string) {
	log.Info("In TestJob")
	for k, v := range jobData {
		valNum, err := strconv.Atoi(v)
		if nil == err {
			jobData[k] = strconv.Itoa(valNum + 1)
		} else {
			log.Error(err)
		}

	}
	sunsetJobData = jobData
}

func (m SunsetJob) JobName() string {
	return "job.TestJob"
}

func (m SunsetJob) GetJobData() map[string]string {
	return sunsetJobData
}
