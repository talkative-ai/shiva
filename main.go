package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/artificial-universe-maker/go-utilities/db"
	"github.com/artificial-universe-maker/go-utilities/prehandle"
	"github.com/artificial-universe-maker/go-utilities/router"
	"github.com/artificial-universe-maker/shiva/routes"
	mux "github.com/gorilla/mux"
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

	doRoute(r, routes.GetProjects)
	doRoute(r, routes.PostProject)
	doRoute(r, routes.PostPublish)
	doRoute(r, routes.PostAuthGoogle)
	doRoute(r, routes.GetProject)
	doRoute(r, routes.GetActor)
	doRoute(r, routes.GetZone)
	doRoute(r, routes.PutActor)
	doRoute(r, routes.PatchProject)

	doRoute(r, routes.DeleteActor)
	doRoute(r, routes.DeleteZone)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://aum.ai", "http://brahman.ngrok.io", "https://brahman.ngrok.io", "https://workbench.aum.ai", "http://localhost:3000", "http://localhost:8080", "http://localhost:3001"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"x-token", "accept", "content-type"},
		ExposedHeaders:   []string{"etag", "x-token"},
		AllowedMethods:   []string{"GET", "PATCH", "POST", "PUT"},
	})

	http.Handle("/", routes.GetIndex.Handler)

	http.Handle("/v1/", c.Handler(r))

	log.Println("Shiva starting server on localhost:8080")

	log.Fatal(http.ListenAndServe(":8080", nil))

}
