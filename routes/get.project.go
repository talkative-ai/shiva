package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"cloud.google.com/go/datastore"

	"github.com/artificial-universe-maker/go-utilities/models"
	"github.com/artificial-universe-maker/go-utilities/myerrors"
	"github.com/artificial-universe-maker/shiva/prehandle"
	"github.com/artificial-universe-maker/shiva/router"

	"strconv"

	mux "github.com/artificial-universe-maker/muxlite"
)

// GetProject router.Route
// Path: "/user/register",
// Method: "GET",
// Accepts models.TokenValidate
// Responds with status of success or failure
var GetProject = &router.Route{
	Path:       "/v1/project/{id:[0-9]+}",
	Method:     "GET",
	Handler:    http.HandlerFunc(getProjectHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT},
}

func getProjectHandler(w http.ResponseWriter, r *http.Request) {
	user := new(models.User)

	urlparams := mux.Vars(r)

	id, err := strconv.ParseInt(urlparams["id"], 10, 64)
	if err != nil {
		myerrors.ServerError(w, r, fmt.Errorf("%v+", urlparams))
		return
	}

	err = json.Unmarshal([]byte(r.Header.Get("X-User")), user)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	dsClient, _ := datastore.NewClient(r.Context(), "artificial-universe-maker")

	project := new(models.AumProject)
	projectKey := datastore.IDKey("Project", id, nil)

	err = dsClient.Get(r.Context(), projectKey, project)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	resp, err := json.Marshal(project)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	fmt.Fprintln(w, string(resp))
}
