package routes

import (
	"encoding/json"
	"log"
	"net/http"

	utilities "github.com/artificial-universe-maker/go-utilities"
	"github.com/artificial-universe-maker/go-utilities/db"
	"github.com/artificial-universe-maker/go-utilities/models"
	"github.com/artificial-universe-maker/go-utilities/router"

	"github.com/artificial-universe-maker/go-utilities/prehandle"
)

// GetProjects router.Route
// Path: "/user/register",
// Method: "GET",
// Accepts models.TokenValidate
// Responds with status of success or failure
var GetProjects = &router.Route{
	Path:       "/v1/projects",
	Method:     "GET",
	Handler:    http.HandlerFunc(getProjectsHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT},
}

func getProjectsHandler(w http.ResponseWriter, r *http.Request) {

	token, err := utilities.ParseJTWClaims(w.Header().Get("x-token"))
	tknData := token["data"].(map[string]interface{})

	// Validate project access
	projects, err := db.DBMap.Select(&models.AumProject{}, `
	SELECT p.* FROM workbench_projects p
	JOIN team_members as m ON m."UserID"=$1
	JOIN teams as t ON t."ID"=m."TeamID"
	WHERE p."TeamID"=t."ID"
	`, tknData["user_id"])
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	results := []map[string]interface{}{}

	for _, project := range projects {
		res := project.(*models.AumProject).PrepareMarshal()
		results = append(results, res)
	}

	json.NewEncoder(w).Encode(results)
}
