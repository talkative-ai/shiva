package prospectacle

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

func init() {

	r := mux.NewRouter()

	r.HandleFunc("/", home)
	r.HandleFunc("/.well-known/acme-challenge/LGBFTrX9DCSCoxEax-Tw36bB0yhJRZoiG2BpbmcM0Ks", ssl)

	http.Handle("/", r)
}

func home(w http.ResponseWriter, rq *http.Request) {
	io.WriteString(w, "Hello World!")
}

func ssl(w http.ResponseWriter, rq *http.Request) {
	io.WriteString(w, "LGBFTrX9DCSCoxEax-Tw36bB0yhJRZoiG2BpbmcM0Ks.xhbKXPDCbpg4pglimVoCtbJVp5X-gqojRN90KtP2Ugc")
}
