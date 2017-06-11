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
	projectKey := datastore.IDKey("Project", project.ID, nil)

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

	for _, location := range project.Locations {
		var k *datastore.Key
		if location.ID != nil {
			k = datastore.IDKey("Location", *location.ID, projectKey)
		} else {
			k = datastore.IncompleteKey("Location", projectKey)
		}

		_, err = dsClient.Put(ctx, k, &location)
		if err != nil {
			myerrors.ServerError(w, r, err)
			return
		}
	}

	for _, object := range project.Objects {
		var k *datastore.Key
		if object.ID != nil {
			k = datastore.IDKey("Object", *object.ID, projectKey)
		} else {
			k = datastore.IncompleteKey("Object", projectKey)
		}

		_, err = dsClient.Put(ctx, k, &object)
		if err != nil {
			myerrors.ServerError(w, r, err)
			return
		}
	}

	for _, npc := range project.NPCs {
		var k *datastore.Key
		if npc.ID != nil {
			k = datastore.IDKey("NPC", *npc.ID, projectKey)
		} else {
			k = datastore.IncompleteKey("NPC", projectKey)
		}

		_, err = dsClient.Put(ctx, k, &npc)
		if err != nil {
			myerrors.ServerError(w, r, err)
			return
		}
	}

	for _, note := range project.Notes {
		var k *datastore.Key
		if note.ID != nil {
			k = datastore.IDKey("Note", *note.ID, projectKey)
		} else {
			k = datastore.IncompleteKey("Note", projectKey)
		}

		_, err = dsClient.Put(ctx, k, &note)
		if err != nil {
			myerrors.ServerError(w, r, err)
			return
		}
	}

	_, err = dsClient.Put(ctx, projectKey, project)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	return
}
