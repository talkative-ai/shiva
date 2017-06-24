package routes

import (
	"encoding/json"
	"net/http"

	"cloud.google.com/go/datastore"

	"github.com/artificial-universe-maker/shiva/models"
	"github.com/artificial-universe-maker/shiva/myerrors"
	"github.com/artificial-universe-maker/shiva/router"

	"github.com/artificial-universe-maker/shiva/prehandle"
)

// PostProject router.Route
// Path: "/user/register",
// Method: "POST",
// Accepts models.TokenValidate
// Responds with status of success or failure
var PostProject = &router.Route{
	Path:       "/v1/project",
	Method:     "POST",
	Handler:    http.HandlerFunc(postProjectHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT, prehandle.RequireBody(65535)},
}

func postProjectHandler(w http.ResponseWriter, r *http.Request) {

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

	project.OwnerID = user.Sub

	dsClient, err := datastore.NewClient(ctx, "artificial-universe-maker")
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	k := datastore.IncompleteKey("Project", nil)

	_, err = dsClient.Put(ctx, k, project)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	return
}
