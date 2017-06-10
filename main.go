package main

import (
	"log"
	"net/http"

	"github.com/rs/cors"
	"github.com/warent/shiva/prehandle"
	"github.com/warent/shiva/router"
	"github.com/warent/shiva/routes"

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

	doRoute(r, routes.GetProjects)
	doRoute(r, routes.PostProject)
	doRoute(r, routes.GetProject)
	doRoute(r, routes.PatchProject)
	doRoute(r, routes.PostProjectLocation)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://aum.ai", "https://workbench.aum.ai", "http://localhost:3000", "http://localhost:3001"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"x-token", "accept", "content-type"},
		ExposedHeaders:   []string{"ETag", "X-Token"},
	})

	http.Handle("/v1/", c.Handler(r))

	// SSL
	http.HandleFunc(routes.GetWellknownAcmeChallenge.Path, routes.GetWellknownAcmeChallenge.Handler.(http.HandlerFunc))

	log.Println("Starting server on localhost:8080")

	appengine.Main()
}
