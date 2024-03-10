package serve

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"go-home/cmd/schedule"
	"go-home/config"

	log "github.com/sirupsen/logrus"
)

func NewServeCmd() (serveCmd *cobra.Command) {
	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start HTTP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			go schedule.StartSchedules()
			return listenAndServe()
		},
	}
	return serveCmd
}

func listenAndServe() error {
	log.Infof("Setting up mux")
	mux := http.NewServeMux()
	mux.HandleFunc("/api/bulb", bulbHandleFunc)

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
	HandleBulbApiRequest(w, r)

}
