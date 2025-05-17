package serve

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/cobra"

	"go-home/cmd/schedule"
	"go-home/config"

	log "github.com/sirupsen/logrus"
)

var disableDb = true

func NewServeCmd() (serveCmd *cobra.Command) {
	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start HTTP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listenAndServe()
		},
	}
	serveCmd.PersistentFlags().BoolVar(&disableDb, "disableDb", true, "Disable state load from SQLite")
	return serveCmd
}

var scheduler gocron.Scheduler

func listenAndServe() error {
	var schedulerChan = make(chan gocron.Scheduler)
	go schedule.StartSchedules(schedulerChan, disableDb)
	log.Infof("Setting up mux")
	mux := http.NewServeMux()
	mux.HandleFunc("/api/bulb", bulbHandleFunc)
	mux.HandleFunc("/api/alarm", alarmHandleFunc)
	mux.HandleFunc("/api/radio", radioHandleFunc)

	scheduler = <-schedulerChan

	log.Infof("Listen and serve")
	return http.ListenAndServe(":"+strconv.Itoa(int(config.ConfigSingleton.Serve.Port)), mux)
}

func parseQueryParamsIntoSlice(s []string) []string {
	var out []string
	for _, element := range s {
		out = append(out, strings.Split(element, ",")...)
	}
	log.Tracef("Processing API call: parsed %v into %v", s, out)
	return out
}

func bulbHandleFunc(w http.ResponseWriter, r *http.Request) {
	log.Infof("Processing bulb API call")
	// /api/bulb?name=bulbName[,bulbName]&op=(on|off)&brightness=0-255&(temperature=2500-6500|color=#RRGGBB|colour=#RRGGBB)
	log.Tracef("1: %v", r.URL.Query()["name"])
	r.URL.Query()["name"] = parseQueryParamsIntoSlice(r.URL.Query()["name"])
	log.Tracef("2: %v", r.URL.Query()["name"])
	handleBulbApiRequest(w, r)
}

func alarmHandleFunc(w http.ResponseWriter, r *http.Request) {
	log.Infof("Processing alarm API call")
	handleAlarmApiRequest(w, r)
}

func radioHandleFunc(w http.ResponseWriter, r *http.Request) {
	log.Infof("Processing alarm API call")
	handleRadioApiRequest(w, r)
}
