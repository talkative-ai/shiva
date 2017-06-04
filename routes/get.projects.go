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

// GetProject router.Route
// Path: "/user/register",
// Method: "GET",
// Accepts models.TokenValidate
// Responds with status of success or failure
var GetProject = &router.Route{
	Path:       "/v1/projects",
	Method:     "GET",
	Handler:    http.HandlerFunc(getProjectHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT},
}

func getProjectHandler(w http.ResponseWriter, r *http.Request) {
	user := new(models.User)

	err := json.Unmarshal([]byte(r.Header.Get("X-User")), user)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	dsClient, _ := datastore.NewClient(r.Context(), "artificial-universe-maker")

	projects := make([]models.Project, 0)

	keys, _ := dsClient.GetAll(r.Context(), datastore.NewQuery("Project").Filter("OwnerID =", user.Sub), &projects)

	for id, _ := range projects {
		projects[id].ID = keys[id].ID
	}

	resp, err := json.Marshal(projects)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	fmt.Fprintln(w, string(resp))
}
