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

// PostProjectLocation router.Route
// Path: "/user/register",
// Method: "POST",
// Accepts models.TokenValidate
// Responds with status of success or failure
var PostProjectLocation = &router.Route{
	Path:       "/v1/project/location",
	Method:     "POST",
	Handler:    http.HandlerFunc(postProjectLocationHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT, prehandle.RequireBody(65535)},
}

func postProjectLocationHandler(w http.ResponseWriter, r *http.Request) {

	type request struct {
		Key       int64                `json:"key"`
		Locations []models.AumLocation `json:"locations"`
	}

	reqBody := new(request)
	user := new(models.User)

	ctx := r.Context()

	err := json.Unmarshal([]byte(r.Header.Get("X-Body")), reqBody)
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

	parentKey := datastore.IDKey("Project", reqBody.Key, nil)

	for _, location := range reqBody.Locations {
		k := datastore.IncompleteKey("Location", parentKey)
		_, err = dsClient.Put(ctx, k, &location)
		if err != nil {
			myerrors.ServerError(w, r, err)
			return
		}
	}

	return
}
