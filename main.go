package prospectacle

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

func init() {

	r := mux.NewRouter()

	r.HandleFunc("/", home)
	r.HandleFunc("/.well-known/acme-challenge/NX1lFw-WUrXPNfjUPvgEmzulsrpsobWGa3SW_GO5MPQ", ssl)

	http.Handle("/", r)
}

func home(w http.ResponseWriter, rq *http.Request) {
	io.WriteString(w, "Hello World!")
}

// [START func_root]
func root(w http.ResponseWriter, rq *http.Request) {

	io.WriteString(w, "NX1lFw-WUrXPNfjUPvgEmzulsrpsobWGa3SW_GO5MPQ.dHP9qyK89IfuTqoierLcfa3E_gZetc7DP7B5SuAznMk")
}

// [END func_root]

func ssl(w http.ResponseWriter, rq *http.Request) {

	io.WriteString(w, "NX1lFw-WUrXPNfjUPvgEmzulsrpsobWGa3SW_GO5MPQ.dHP9qyK89IfuTqoierLcfa3E_gZetc7DP7B5SuAznMk")
}
