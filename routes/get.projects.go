package routes

import (
	"encoding/json"
	"net/http"

	"github.com/artificial-universe-maker/go-utilities/models"
	"github.com/artificial-universe-maker/go-utilities/myerrors"
	"github.com/artificial-universe-maker/shiva/router"

	"github.com/artificial-universe-maker/shiva/prehandle"
)

// GetProjects router.Route
// Path: "/user/register",
// Method: "GET",
// Accepts models.TokenValidate
// Responds with status of success or failure
var GetProjects = &router.Route{
	Path:       "/v1/projects",
	Method:     "GET",
	Handler:    http.HandlerFunc(getProjectsHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT},
}

func getProjectsHandler(w http.ResponseWriter, r *http.Request) {
	user := new(models.User)

	err := json.Unmarshal([]byte(r.Header.Get("X-User")), user)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}
}
