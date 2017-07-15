package routes

import (
	"net/http"

	"github.com/artificial-universe-maker/shiva/router"

	"github.com/artificial-universe-maker/shiva/prehandle"
)

// PatchProject router.Route
// Path: "/user/register",
// Method: "PATCH",
// Accepts models.TokenValidate
// Responds with status of success or failure
var PatchProjects = &router.Route{
	Path:       "/v1/projects",
	Method:     "PATCH",
	Handler:    http.HandlerFunc(patchProjectsHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT, prehandle.RequireBody(65535)},
}

func patchProjectsHandler(w http.ResponseWriter, r *http.Request) {

}
