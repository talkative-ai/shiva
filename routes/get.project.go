package routes

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/artificial-universe-maker/go-utilities/db"
	"github.com/artificial-universe-maker/go-utilities/models"
	"github.com/artificial-universe-maker/go-utilities/myerrors"
	"github.com/artificial-universe-maker/go-utilities/prehandle"
	"github.com/artificial-universe-maker/go-utilities/router"

	mux "github.com/artificial-universe-maker/muxlite"
)

// GetProject router.Route
// Path: "/project/{id}",
// Method: "GET",
// Accepts models.TokenValidate
// Responds with the project data
var GetProject = &router.Route{
	Path:       "/v1/project/{id:[0-9]+}",
	Method:     "GET",
	Handler:    http.HandlerFunc(getProjectHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT},
}

func getProjectHandler(w http.ResponseWriter, r *http.Request) {

	urlparams := mux.Vars(r)

	id, err := strconv.ParseInt(urlparams["id"], 10, 64)
	if err != nil {
		myerrors.Respond(w, &myerrors.MySimpleError{
			Code:    http.StatusBadRequest,
			Message: "bad_id",
			Req:     r,
		})
		return
	}

	// Validate project access
	project := &models.AumProject{}
	err = db.DBMap.SelectOne(project, "SELECT * FROM workbench_projects WHERE id=$1", id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	member := &models.TeamMember{}
	err = db.DBMap.SelectOne(member, `SELECT t."Role" FROM workbench_projects AS p JOIN team_members AS t ON t."TeamID"=p."TeamID" AND t."UserID"=1 WHERE p."ID"=1`)
	if member.TeamID != 1 || err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Return project data
	json.NewEncoder(w).Encode(project)
}
