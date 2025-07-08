package serve

import (
	"net/http"
	"time"

	"go-home/cmd/schedule"

	"github.com/go-co-op/gocron/v2"
	log "github.com/sirupsen/logrus"
)

func handleAlarmApiRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		log.Tracef("Processing API call: GET")
		return

	case http.MethodDelete:
		log.Tracef("Processing API call: Delete")

		switch r.URL.Query().Get("op") {
		case "nextWakeup":
			log.Tracef("Processing API call: delete next wakeup alarm")
			scheduler.RemoveByTags("OPNAME:wakeup")
			jobName := "builtin_postponewakeup"
			scheduler.RemoveByTags(jobName)
			_, err := scheduler.NewJob(
				gocron.OneTimeJob(
					gocron.OneTimeJobStartDateTime(time.Now().AddDate(0, 0, 1))),
				gocron.NewTask(func() {
					// reload all jobs from db including wakeup
					schedule.RestartSchedules(scheduler)
				}),
				gocron.WithName(jobName),
				gocron.WithTags(jobName))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		case "until":
			log.Tracef("Processing API call: delete alarms until")
			tmpNow := time.Now()
			until := r.URL.Query().Get("until")
			parsedUntil, err := time.Parse("2006-01-02", until)
			if err != nil {
				log.Errorf("Processing API call: cannot parse until date")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			updatedUntil := time.Date(
				parsedUntil.Year(),
				parsedUntil.Month(),
				parsedUntil.Day(),
				tmpNow.Hour(),
				tmpNow.Minute(),
				tmpNow.Second(),
				tmpNow.Nanosecond(),
				tmpNow.Location(),
			)
			scheduler.RemoveByTags("OPNAME:wakeup")
			jobName := "builtin_postponewakeup"
			scheduler.RemoveByTags(jobName)
			_, err = scheduler.NewJob(
				gocron.OneTimeJob(
					gocron.OneTimeJobStartDateTime(updatedUntil)),
				gocron.NewTask(func() {
					// reload all jobs from db including wakeup
					schedule.RestartSchedules(scheduler)
				}),
				gocron.WithName(jobName),
				gocron.WithTags(jobName))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		default:
			log.Errorf("Processing API call: unknown operation")
			w.WriteHeader(http.StatusBadRequest)
		}

	case http.MethodOptions:
		log.Tracef("Processing API call: OPTIONS")
		w.Header().Set("Allow", "GET, DELETE, OPTIONS")
		w.WriteHeader(http.StatusNoContent)
		return

	default:
		log.Tracef("Processing API call: method missing")
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return

	}
}
