package routes

import (
	"encoding/json"
	"net/http"

	utilities "github.com/artificial-universe-maker/core"
	"github.com/artificial-universe-maker/core/db"
	"github.com/artificial-universe-maker/core/models"
	"github.com/artificial-universe-maker/core/myerrors"
	"github.com/artificial-universe-maker/core/router"
	uuid "github.com/artificial-universe-maker/go.uuid"

	"github.com/artificial-universe-maker/core/prehandle"
)

// GetProjects router.Route
// Path: "/user/register",
// Method: "GET",
// Accepts models.TokenValidate
// Responds with status of success or failure
var GetProjects = &router.Route{
	Path:       "/workbench/v1/projects",
	Method:     "GET",
	Handler:    http.HandlerFunc(getProjectsHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT},
}

func getProjectsHandler(w http.ResponseWriter, r *http.Request) {

	token, err := utilities.ParseJTWClaims(w.Header().Get("x-token"))
	tknData := token["data"].(map[string]interface{})
	userID, err := uuid.FromString(tknData["user_id"].(string))
	if err != nil {
		myerrors.Respond(w, &myerrors.MySimpleError{
			Code:    http.StatusBadRequest,
			Message: "bad_id",
			Req:     r,
		})
		return
	}

	// Validate project access
	projects, err := db.DBMap.Select(&models.AumProject{}, `
	SELECT p.* FROM workbench_projects p
	JOIN team_members as m ON m."UserID"=$1
	JOIN teams as t ON t."ID"=m."TeamID"
	WHERE p."TeamID"=t."ID"
	`, userID)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	results := []map[string]interface{}{}

	for _, project := range projects {
		res := project.(*models.AumProject).PrepareMarshal()
		results = append(results, res)
	}

	json.NewEncoder(w).Encode(results)
}
