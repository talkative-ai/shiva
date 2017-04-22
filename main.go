package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/warent/phrhero-backend/prehandle"
	"github.com/warent/phrhero-backend/router"
	"github.com/warent/phrhero-backend/routes"

	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

func doRoute(r *mux.Router, route *router.Route) {
	r.Handle(route.Path, prehandle.PreHandle(route.Handler.(http.HandlerFunc), route.Prehandler...)).Methods(route.Method)
}

func main() {
	r := mux.NewRouter()
	doRoute(r, routes.PostUserRegister)

	http.Handle("/", r)
}

func create(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	vars := mux.Vars(r)

	k := datastore.NewKey(ctx, "Entity", "stringID", 0, nil)
	e := new(Entity)
	if err := datastore.Get(ctx, k, e); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	old := e.Value
	e.Value = vars["value"]

	if _, err := datastore.Put(ctx, k, e); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "old=%q\nnew=%q\n", old, e.Value)
}

func home(w http.ResponseWriter, rq *http.Request) {
	io.WriteString(w, "Hello World!")
}

func ssl(w http.ResponseWriter, rq *http.Request) {
	io.WriteString(w, "LGBFTrX9DCSCoxEax-Tw36bB0yhJRZoiG2BpbmcM0Ks.xhbKXPDCbpg4pglimVoCtbJVp5X-gqojRN90KtP2Ugc")
}
