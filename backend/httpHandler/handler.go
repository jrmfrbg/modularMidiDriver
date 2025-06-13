package httphandler

import (
	"fmt"
	"net/http"
)

// Route defines a mapping between a URL path and its handler function.
type Route struct {
	Path    string
	Handler http.HandlerFunc
}

var TestCallRoute = Route{
	Path: "/testCall",
	Handler: func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello Client")
	},
}

// StartHTTPServer starts an HTTP server on the given port and uses the provided routes for GET requests.
func StartHTTPServer(port string, routes []Route) error {
	mux := http.NewServeMux()
	for _, route := range routes {
		// Wrap each handler to only allow GET requests
		mux.HandleFunc(route.Path, func(handler http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					handler(w, r)
				} else {
					http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				}
			}
		}(route.Handler))
	}
	addr := fmt.Sprintf(":%s", port)
	return http.ListenAndServe(addr, mux)
}
