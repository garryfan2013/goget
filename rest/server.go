package rest

import (
	"net/http"

	"github.com/garryfan2013/goget/proxy"
	"github.com/garryfan2013/goget/rest/route"
)

type RestServer struct {
	pm proxy.ProxyManager
}

func NewRestServer(pm proxy.ProxyManager) (*RestServer, error) {
	return &RestServer{pm}, nil
}

func (s *RestServer) Serve() error {
	router := route.NewRouter(s)
	return http.ListenAndServe("0.0.0.0:8000", router.Handler())
}

// Implementation of HandlerWrapper interface
func (s *RestServer) Wrap(f route.RouteHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f(w, r, s.pm)
	}
}
