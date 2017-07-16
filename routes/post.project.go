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
	//user := new(models.User)

	// err := json.Unmarshal([]byte(r.Header.Get("X-Body")), user)
	// if err != nil {
	// 	myerrors.ServerError(w, r, err)
	// 	return
	// }

	// fmt.Println("Here we are", r.Header.Get("X-Body"))

	err := json.Unmarshal([]byte(r.Header.Get("X-Body")), project)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	// project.OwnerID = user.Sub

	err = db.DBMap.Insert(project)
	if err != nil {
		fmt.Println(err)
	}

	return
}
