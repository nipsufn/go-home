package serve

import (
	"net/http"
	"strconv"

	"github.com/spf13/cobra"

	"go-home/cmd/schedule"
	"go-home/config"
)

func NewServeCmd() (serveCmd *cobra.Command) {
	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start HTTP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			go schedule.StartSchedules()
			go schedule.MasterBulbLoop()
			return listenAndServe()
		},
	}
	return serveCmd
}

func listenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api", api)

	return http.ListenAndServe(":"+strconv.Itoa(int(config.ConfigSingleton.Serve.Port)), mux)
}

func api(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api" {
		http.NotFound(w, r)
		return
	}

	if r.URL.Query().Has("bulb") {

	}

	switch r.Method {
	case http.MethodGet:
		// Handle the GET request...

	case http.MethodPost:
		// Handle the POST request...

	case http.MethodOptions:
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		w.WriteHeader(http.StatusNoContent)

	default:
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
