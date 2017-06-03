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

// PostProject router.Route
// Path: "/user/register",
// Method: "POST",
// Accepts models.TokenValidate
// Responds with status of success or failure
var PostProject = &router.Route{
	Path:       "/v1/project",
	Method:     "POST",
	Handler:    http.HandlerFunc(postProjectHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT, prehandle.RequireBody(5120)},
}

func postProjectHandler(w http.ResponseWriter, r *http.Request) {

	project := new(models.Project)
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
