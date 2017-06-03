package main

import (
	"net/http"

	"github.com/rs/cors"
	"github.com/warent/shiva/routes"
	"github.com/warent/shiva/prehandle"
	"github.com/warent/shiva/router"

	"github.com/gorilla/mux"

	"google.golang.org/appengine"
)

func doRoute(r *mux.Router, route *router.Route) {
	r.Handle(route.Path, prehandle.PreHandle(route.Handler.(http.HandlerFunc), route.Prehandler...)).Methods(route.Method)
}

func main() {

	r := mux.NewRouter()
	doRoute(r, routes.GetIndex)

	doRoute(r, routes.PostTokenValidate)

	doRoute(r, routes.PostProject)
	doRoute(r, routes.GetProject)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://aum.ai", "https://dev.aum.ai", "http://localhost:3000"},
		AllowCredentials: true,
	})

	// Insert the middleware
	http.Handle("/v1/", c.Handler(r))

	// SSL
	http.HandleFunc(routes.GetWellknownAcmeChallenge.Path, routes.GetWellknownAcmeChallenge.Handler.(http.HandlerFunc))

	appengine.Main()
}
