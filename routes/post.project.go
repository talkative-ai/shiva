package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/artificial-universe-maker/shiva/db"
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
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.RequireBody(65535)},
}

func postProjectHandler(w http.ResponseWriter, r *http.Request) {

	project := new(models.AumProject)

	err := json.Unmarshal([]byte(r.Header.Get("X-Body")), project)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}
	_, err = db.ShivaDB.NamedExec("INSERT INTO projects (title) VALUES (:title)", project)
	if err != nil {
		fmt.Println(err)
	}

	// err = json.Unmarshal([]byte(r.Header.Get("X-User")), user)
	// if err != nil {
	// 	myerrors.ServerError(w, r, err)
	// 	return
	// }

	return
}
