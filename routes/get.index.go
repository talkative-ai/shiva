package routes

import (
	"fmt"
	"net/http"

	"github.com/talkative-ai/core/prehandle"
	"github.com/talkative-ai/core/router"
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
	fmt.Fprintf(w, "Want to help us build this api and more? Email me: wyatt@talkative.ai")
}
