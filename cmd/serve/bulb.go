package serve

import (
	"go-home/cmd/bulb"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func handleBulbApiRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		log.Tracef("Processing API call: GET")
		return

	case http.MethodPost:
		log.Tracef("Processing API call: POST")
		q := r.URL.Query()
		log.Tracef("3: %v", q["name"])
		if len(q["name"]) == 0 {
			q["name"] = append(q["name"], "all")
		}

		log.Tracef("4: %v", q["name"])
		if len(q["op"]) != 1 {
			log.Errorf("Processing API call: parameter `op` duplicated")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		switch q.Get("op") {
		case "off":
			log.Tracef("Processing API call: op off on bulbs %v", q["name"])
			bulb.TurnBulbOffByName(q["name"]...)
			w.WriteHeader(http.StatusOK)
			return
		case "on":
			log.Tracef("Processing API call: op on on bulbs %v", q["name"])
			if len(q["brightness"]) != 1 || (len(q["temperature"])+len(q["colour"])+len(q["color"]) != 1) {
				log.Errorf("Processing API call: params invalid")
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			var (
				brightness  uint64
				temperature int
				color       string
				err         error
			)
			color = q.Get("colour") + q.Get("color")
			if brightness, err = strconv.ParseUint(q.Get("brightness"), 10, 8); err != nil {
				log.Errorf("Processing API call: cannot cast brightness as uint8")
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if len(color) == 0 {
				if temperature, err = strconv.Atoi(q.Get("temperature")); err != nil {
					log.Errorf("Processing API call: cannot cast temperature as int")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			} else {
				temperature = 0
			}
			err = bulb.TurnBulbOnByName(uint8(brightness), uint(temperature), color, q["name"]...)
			if err != nil {
				log.Errorf("Processing API call: failed to turn bulb on: %v", err)
				w.Write([]byte("Failed to turn bulb on"))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
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
