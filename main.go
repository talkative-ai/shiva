package main

import (
	"log"
	"net/http"

	"github.com/artificial-universe-maker/core/router"
	"github.com/artificial-universe-maker/shiva/routes"
	mux "github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {

	r := mux.NewRouter()
	router.ApplyRoute(r, routes.GetIndex)

	router.ApplyRoute(r, routes.GetProjects)
	router.ApplyRoute(r, routes.PostProject)
	router.ApplyRoute(r, routes.PostPublish)
	router.ApplyRoute(r, routes.PostAuthGoogle)
	router.ApplyRoute(r, routes.GetProject)
	router.ApplyRoute(r, routes.GetProjectMetadata)
	router.ApplyRoute(r, routes.GetActor)
	router.ApplyRoute(r, routes.GetZone)
	router.ApplyRoute(r, routes.PatchActor)
	router.ApplyRoute(r, routes.PatchProject)

	router.ApplyRoute(r, routes.DeleteActor)
	router.ApplyRoute(r, routes.DeleteZone)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://aum.ai", "https://harihara.ngrok.io", "http://brahman.ngrok.io", "https://brahman.ngrok.io", "https://workbench.aum.ai", "http://localhost:3000", "http://localhost:8080", "http://localhost:3001"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"x-token", "accept", "content-type"},
		ExposedHeaders:   []string{"etag", "x-token"},
		AllowedMethods:   []string{"GET", "PATCH", "POST", "PUT"},
	})

	r.Handle("/healthz", routes.GetIndex.Handler)
	http.Handle("/", c.Handler(r))

	log.Println("Shiva starting server on localhost:8080")

	log.Fatal(http.ListenAndServe(":8080", nil))

}
