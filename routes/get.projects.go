package routes

import (
	"encoding/json"
	"net/http"

	utilities "github.com/talkative-ai/core"
	"github.com/talkative-ai/core/db"
	"github.com/talkative-ai/core/models"
	"github.com/talkative-ai/core/myerrors"
	"github.com/talkative-ai/core/router"
	uuid "github.com/talkative-ai/go.uuid"

	"github.com/talkative-ai/core/prehandle"
)

// GetProjects router.Route
// Path: "/projects"
// Method: "GET"
// Responds with an array of projects
var GetProjects = &router.Route{
	Path:       "/workbench/v1/projects",
	Method:     "GET",
	Handler:    http.HandlerFunc(getProjectsHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT},
}

func getProjectsHandler(w http.ResponseWriter, r *http.Request) {

	// Parse the token and token data
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

	// Validate access permissions to the project
	// while simultaneously fetching project data
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

	// Map all of the results to an array of JSON objects
	results := []map[string]interface{}{}
	for _, project := range projects {
		res := project.(*models.AumProject).PrepareMarshal()
		results = append(results, res)
	}

	// Return the projects
	json.NewEncoder(w).Encode(results)
}
