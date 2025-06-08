package serve

import (
	"go-home/cmd/playback"
	"go-home/config"
	"net/http"
	"net/url"
	"time"

	log "github.com/sirupsen/logrus"
)

func handleRadioApiRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		log.Tracef("Processing API call: GET")
		return

	case http.MethodPost:
		log.Tracef("Processing API call: POST")

		switch r.URL.Query().Get("op") {
		case "on":
			station := r.URL.Query().Get("station")
			if station != "" {
				station = config.ConfigSingleton.Radio.DefaultStation
			}
			stationUrl := config.ConfigSingleton.Radio.Stations[station]
			log.Tracef("Playing radio %s: %s", station, stationUrl)
			if err := playback.PlayURL(url.URL(*stationUrl), time.Duration(config.ConfigSingleton.Radio.FadeDelaySec)*time.Second); err != nil {
				log.Errorf("Playback error: %v", err)
			}
			return
		case "off":
			playback.Clear(time.Duration(config.ConfigSingleton.Radio.FadeDelaySec) * time.Second)
			return
		default:
			log.Errorf("Processing API call")
			w.WriteHeader(http.StatusBadRequest)
		}

	case http.MethodOptions:
		log.Tracef("Processing API call: OPTIONS")
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		w.WriteHeader(http.StatusNoContent)
		return

	default:
		log.Tracef("Processing API call: method missing")
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return

	}
}
