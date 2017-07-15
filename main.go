package main

import (
	"fmt"
	"log"
	"net/http"

	mux "github.com/artificial-universe-maker/muxlite"
	"github.com/artificial-universe-maker/shiva/db"
	"github.com/artificial-universe-maker/shiva/prehandle"
	"github.com/artificial-universe-maker/shiva/router"
	"github.com/artificial-universe-maker/shiva/routes"
	"github.com/rs/cors"
)

func doRoute(r *mux.Router, route *router.Route) {
	r.Handle(route.Path, prehandle.PreHandle(route.Handler.(http.HandlerFunc), route.Prehandler...)).Methods(route.Method)
}

func main() {

	err := db.InitializeDB()
	if err != nil {
		fmt.Println(err)
		return
	}

	r := mux.NewRouter()
	doRoute(r, routes.GetIndex)

	doRoute(r, routes.PostTokenValidate)

	doRoute(r, routes.GetProjects)
	doRoute(r, routes.PostProject)
	doRoute(r, routes.GetProject)
	doRoute(r, routes.PatchProjects)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://aum.ai", "https://workbench.aum.ai", "http://localhost:3000", "http://localhost:3001"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"x-token", "accept", "content-type"},
		ExposedHeaders:   []string{"ETag", "X-Token"},
		AllowedMethods:   []string{"GET", "PATCH", "POST"},
	})

	http.Handle("/v1/", c.Handler(r))

	// SSL
	http.HandleFunc(routes.GetWellknownAcmeChallenge.Path, routes.GetWellknownAcmeChallenge.Handler.(http.HandlerFunc))

	log.Println("Starting server on localhost:8080")

	http.ListenAndServe(":8080", nil)
}
