package prospectacle

import (
	"fmt"
	"io"
	"net"
	"net/http"

	"os"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/socket"
)

func init() {

	r := mux.NewRouter()

	r.HandleFunc("/", home)
	r.HandleFunc("/create/{value}", create)
	r.HandleFunc("/.well-known/acme-challenge/LGBFTrX9DCSCoxEax-Tw36bB0yhJRZoiG2BpbmcM0Ks", ssl)
	r.HandleFunc("/redis/{value}", writeToRedis)

	http.Handle("/", r)
}

func writeToRedis(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ctx := appengine.NewContext(r)

	client := redis.NewClient(&redis.Options{
		Dialer: func() (net.Conn, error) {
			return socket.Dial(ctx, "tcp", os.Getenv("REDIS_ADDR"))
		},
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	pong, err := client.Ping().Result()
	if err != nil {
		log.Criticalf(ctx, err.Error())
		return
	}

	log.Debugf(ctx, pong)
	client.Set("Saved", vars["value"], -1)
}

type Entity struct {
	Value string
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
