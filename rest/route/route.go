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
	Handler RouteHandlerFunc
}

type Router struct {
	r *mux.Router
}

type Routes []Route

type HandlerWrapper interface {
	Wrap(RouteHandlerFunc) http.HandlerFunc
}

type RouteHandlerFunc func(http.ResponseWriter, *http.Request, interface{})

func register(r *Route) {
	routes[r.Name] = r
}

func NewRouter(wrapper HandlerWrapper) *Router {
	router := &Router{
		r: mux.NewRouter(),
	}

	for _, route := range routes {
		router.r.Methods(route.Method).Path(route.Pattern).Name(route.Name).Handler(wrapper.Wrap(route.Handler))
	}

	return router
}

func (r *Router) Handler() *mux.Router {
	return r.r
}
