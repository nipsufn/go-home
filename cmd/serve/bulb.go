package serve

import (
	"go-home/cmd/bulb"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func HandleBulbApiRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		log.Tracef("Processing API call: GET")
		return

	case http.MethodPost:
		log.Tracef("Processing API call: POST")

		log.Tracef("3: %v", r.URL.Query()["name"])
		if len(r.URL.Query()["name"]) == 0 {
			r.URL.Query()["name"] = append(r.URL.Query()["name"], "all")
		}

		log.Tracef("4: %v", r.URL.Query()["name"])
		if len(r.URL.Query()["op"]) != 1 {
			log.Errorf("Processing API call: parameter `op` duplicated")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		switch r.URL.Query().Get("op") {
		case "off":
			log.Tracef("Processing API call: op off on bulbs %v", r.URL.Query()["name"])
			bulb.TurnBulbOffByName(r.URL.Query()["name"]...)
			w.WriteHeader(http.StatusOK)
			return
		case "on":
			log.Tracef("Processing API call: op on on bulbs %v", r.URL.Query()["name"])
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
			bulb.TurnBulbOnByName(uint8(brightness), uint(temperature), color, r.URL.Query()["name"]...)
			return
		default:
			log.Errorf("Processing API call: parameter `op` has to be either `on` or `off`")
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
