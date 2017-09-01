package routes

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/artificial-universe-maker/go-utilities/myerrors"
	"github.com/artificial-universe-maker/go-utilities/prehandle"
	"github.com/artificial-universe-maker/go-utilities/router"

	mux "github.com/artificial-universe-maker/muxlite"
)

// GetProject router.Route
// Path: "/project/{id}",
// Method: "GET",
// Accepts models.TokenValidate
// Responds with the project data
var GetProject = &router.Route{
	Path:       "/v1/project/{id:[0-9]+}",
	Method:     "GET",
	Handler:    http.HandlerFunc(getProjectHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT},
}

func getProjectHandler(w http.ResponseWriter, r *http.Request) {

	urlparams := mux.Vars(r)

	id, err := strconv.ParseInt(urlparams["id"], 10, 64)
	if err != nil {
		myerrors.ServerError(w, r, fmt.Errorf("%v+", urlparams))
		return
	}

	// Validate token

	// Validate project access

	// Return project data

}
