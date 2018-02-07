package routes

import (
	"fmt"
	"net/http"

	"github.com/artificial-universe-maker/core/prehandle"
	"github.com/artificial-universe-maker/core/router"
)

// GetIndex router.Route
/* Path: "/"
 * Method: "GET"
 * Responds with status of success or failure
 */
var GetIndex = &router.Route{
	Path:       "/workbench/v1/",
	Method:     "GET",
	Handler:    http.HandlerFunc(getIndexHandler),
	Prehandler: []prehandle.Prehandler{},
}

func getIndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Want to help us build this api and more? Email us: info@aum.ai")
}
