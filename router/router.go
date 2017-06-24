package router

import (
	"net/http"

	"github.com/artificial-universe-maker/shiva/prehandle"

	mux "github.com/artificial-universe-maker/muxlite"
)

// Route contains route information for multiplexing
type Route struct {
	http.Handler
	Prehandler []prehandle.Prehandler
	Method     mux.Method
	Path       string
}

// Test ceates a new mux.Router instance for easy testing. This additionally allows support for Mux params
func (route *Route) Test(w http.ResponseWriter, r *http.Request) {
	m := mux.NewRouter()
	m.Handle(route.Path, prehandle.PreHandle(route.Handler.(http.HandlerFunc), route.Prehandler...)).Methods(route.Method)
	m.ServeHTTP(w, r)
}
