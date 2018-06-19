package route

import (
	"net/http"

	"github.com/gorilla/mux"
)

var (
	routes map[string]*Route = make(map[string]*Route)
)

type Route struct {
	Name    string
	Method  string
	Pattern string
	Handler http.HandlerFunc
}

type Routes []Route

func register(r *Route) {
	routes[r.Name] = r
}

func NewRouter() *mux.Router {
	router := mux.NewRouter()
	for _, r := range routes {
		router.Methods(r.Method).Path(r.Pattern).Name(r.Name).Handler(r.Handler)
	}

	return router
}
