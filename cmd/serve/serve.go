package serve

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"go-home/cmd/bulb"
	"go-home/config"

	log "github.com/sirupsen/logrus"
)

func NewServeCmd() (serveCmd *cobra.Command) {
	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start HTTP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			//go schedule.StartSchedules()
			//go schedule.MasterBulbLoop()
			return listenAndServe()
		},
	}
	return serveCmd
}

func listenAndServe() error {
	log.Infof("Setting up mux")
	mux := http.NewServeMux()
	mux.HandleFunc("/api", api)

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

func api(w http.ResponseWriter, r *http.Request) {
	log.Infof("Processing API call")
	if r.URL.Path != "/api" {
		log.Tracef("Processing API call: wrong path")
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		log.Tracef("Processing API call: GET")

	case http.MethodPost:
		log.Tracef("Processing API call: POST")
		if r.URL.Query().Has("bulb") {
			log.Tracef("Processing API call: param bulb present")
			log.Tracef("Processing API call: param op: %v", r.URL.Query()["op"])
			if len(r.URL.Query()["op"]) == 1 {
				if r.URL.Query().Get("op") == "off" {
					log.Tracef("Processing API call: op off on bulbs %v", r.URL.Query()["bulb"])
					bulb.TurnBulbOffByName(parseQueryParamsIntoSlice(r.URL.Query()["bulb"])...)
					w.WriteHeader(http.StatusOK)
				} else if r.URL.Query().Get("op") == "on" {
					log.Tracef("Processing API call: op on on bulbs %v", r.URL.Query()["bulb"])
					if len(r.URL.Query()["brightness"]) != 1 && (len(r.URL.Query()["temperature"])+len(r.URL.Query()["colour"])+len(r.URL.Query()["color"]) != 1) {
						log.Errorf("Processing API call: params invalid")
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					var (
						brightness  int
						temperature int
						color       string
						err         error
					)
					color = r.URL.Query().Get("colour") + r.URL.Query().Get("color")
					if brightness, err = strconv.Atoi(r.URL.Query().Get("brightness")); err != nil {
						log.Errorf("Processing API call: cannot cast brightness as int")
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					if temperature, err = strconv.Atoi(r.URL.Query().Get("temperature")); err != nil {
						log.Errorf("Processing API call: cannot cast temperature as int")
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					bulb.TurnBulbOnByName(uint8(brightness), uint(temperature), color, parseQueryParamsIntoSlice(r.URL.Query()["bulb"])...)
				} else {
					log.Errorf("Processing API call: parameter `op` has to be either `on` or `off`")
					w.WriteHeader(http.StatusBadRequest)
				}
			} else {
				log.Errorf("Processing API call: parameter `op` duplicated")
				w.WriteHeader(http.StatusBadRequest)
			}

		}

	case http.MethodOptions:
		log.Tracef("Processing API call: OPTIONS")
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		w.WriteHeader(http.StatusNoContent)

	default:
		log.Tracef("Processing API call: method missing")
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
