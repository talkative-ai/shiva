package routes

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/artificial-universe-maker/go-utilities/prehandle"
	"github.com/artificial-universe-maker/go-utilities/router"
)

// GetIndex router.Route
// Path: "/user/register",
// Method: "POST",
// Accepts models.UserRegister
// Responds with status of success or failure
var GetIndex = &router.Route{
	Path:       "/workbench/v1/",
	Method:     "GET",
	Handler:    http.HandlerFunc(getIndexHandler),
	Prehandler: []prehandle.Prehandler{},
}

func getIndexHandler(w http.ResponseWriter, r *http.Request) {
	bs, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println("Error", err)
	}
	fmt.Println(string(bs))
	fmt.Fprintf(w, "Want to help us build this api and more? Email us: dev@aum.ai")

}
