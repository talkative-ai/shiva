package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"cloud.google.com/go/datastore"

	"github.com/warent/shiva/models"
	"github.com/warent/shiva/myerrors"
	"github.com/warent/shiva/router"

	"strconv"

	"github.com/gorilla/mux"
	"github.com/warent/shiva/prehandle"
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

	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)

	err := json.Unmarshal([]byte(r.Header.Get("X-User")), user)
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

	project.ID = id

	locations := make([]models.AumLocation, 0)
	keys, _ := dsClient.GetAll(r.Context(), datastore.NewQuery("Location").Ancestor(projectKey), &locations)
	for id := range locations {
		locations[id].ID = &keys[id].ID
	}
	project.Locations = locations

	objects := make([]models.AumObject, 0)
	keys, _ = dsClient.GetAll(r.Context(), datastore.NewQuery("Object").Ancestor(projectKey), &objects)
	for id := range objects {
		objects[id].ID = &keys[id].ID
	}
	project.Objects = objects

	npcs := make([]models.AumNPC, 0)
	keys, _ = dsClient.GetAll(r.Context(), datastore.NewQuery("NPC").Ancestor(projectKey), &npcs)
	for id := range npcs {
		npcs[id].ID = &keys[id].ID
	}
	project.NPCs = npcs

	Notes := make([]models.AumNote, 0)
	keys, _ = dsClient.GetAll(r.Context(), datastore.NewQuery("Note").Ancestor(projectKey), &Notes)
	for id := range Notes {
		Notes[id].ID = &keys[id].ID
	}
	project.Notes = Notes

	resp, err := json.Marshal(project)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	fmt.Fprintln(w, string(resp))
}