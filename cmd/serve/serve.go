package serve

import (
	"net/http"

	"github.com/spf13/cobra"

	"go-home/cmd/schedule"
)

func NewServeCmd() (serveCmd *cobra.Command) {
	serveCmd = &cobra.Command{
		Use:   "list",
		Short: "Turn lightbulb(s) on",
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
	mux.HandleFunc("/", index)

	return http.ListenAndServe(":3000", mux)

}

func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Common code for all requests can go here...

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
