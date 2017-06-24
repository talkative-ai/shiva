package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"cloud.google.com/go/datastore"

	"github.com/warent/shiva/models"
	"github.com/warent/shiva/myerrors"
	"github.com/warent/shiva/router"

	"github.com/warent/shiva/prehandle"
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

	dsClient, err := datastore.NewClient(r.Context(), "artificial-universe-maker")
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	projects := make([]models.AumProject, 0)

	keys, err := dsClient.GetAll(r.Context(), datastore.NewQuery("Project").Filter("OwnerID =", user.Sub), &projects)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	for id := range projects {
		projects[id].ID = keys[id].ID
	}

	resp, err := json.Marshal(projects)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	fmt.Fprintln(w, string(resp))
}
