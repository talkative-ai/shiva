package routes

import (
	"encoding/json"
	"net/http"

	"cloud.google.com/go/datastore"

	"github.com/warent/shiva/models"
	"github.com/warent/shiva/myerrors"
	"github.com/warent/shiva/router"

	"github.com/warent/shiva/prehandle"
)

// PatchProject router.Route
// Path: "/user/register",
// Method: "PATCH",
// Accepts models.TokenValidate
// Responds with status of success or failure
var PatchProject = &router.Route{
	Path:       "/v1/project",
	Method:     "PATCH",
	Handler:    http.HandlerFunc(patchProjectHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT, prehandle.RequireBody(65535)},
}

func patchProjectHandler(w http.ResponseWriter, r *http.Request) {

	project := new(models.AumProject)
	user := new(models.User)

	ctx := r.Context()

	err := json.Unmarshal([]byte(r.Header.Get("X-Body")), project)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	err = json.Unmarshal([]byte(r.Header.Get("X-User")), user)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	dsClient, err := datastore.NewClient(ctx, "artificial-universe-maker")
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	k := datastore.IDKey("Project", project.ID, nil)

	_, err = dsClient.Put(ctx, k, project)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	return
}
