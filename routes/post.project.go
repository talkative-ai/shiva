package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/artificial-universe-maker/go-utilities/db"
	"github.com/artificial-universe-maker/go-utilities/models"
	"github.com/artificial-universe-maker/go-utilities/myerrors"
	"github.com/artificial-universe-maker/go-utilities/prehandle"
	"github.com/artificial-universe-maker/go-utilities/router"
)

// PostProject router.Route
// Path: "/user/register",
// Method: "POST",
// Accepts models.TokenValidate
// Responds with status of success or failure
var PostProject = &router.Route{
	Path:       "/workbench/v1/project",
	Method:     "POST",
	Handler:    http.HandlerFunc(postProjectHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT, prehandle.RequireBody(65535)},
}

type postProjectRequest struct {
	Title string
}

func postProjectHandler(w http.ResponseWriter, r *http.Request) {

	project := new(models.AumProject)
	postProject := postProjectRequest{}

	err := json.Unmarshal([]byte(r.Header.Get("X-Body")), postProject)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	project.Title = postProject.Title

	err = db.DBMap.Insert(project)
	if err != nil {
		fmt.Println(err)
	}

	return
}
