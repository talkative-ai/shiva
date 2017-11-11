package routes

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	utilities "github.com/artificial-universe-maker/core"
	"github.com/artificial-universe-maker/core/db"
	"github.com/artificial-universe-maker/core/models"
	"github.com/artificial-universe-maker/core/myerrors"
	"github.com/artificial-universe-maker/core/prehandle"
	"github.com/artificial-universe-maker/core/router"
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

	token, err := utilities.ParseJTWClaims(w.Header().Get("x-token"))
	tknData := token["data"].(map[string]interface{})

	project := new(models.AumProject)
	postProject := postProjectRequest{}

	err = json.Unmarshal([]byte(r.Header.Get("X-Body")), &postProject)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	var count int

	res, err := db.Instance.Query(`
		SELECT COUNT(p."ID") as "ProjectCount"
		FROM workbench_projects p
		JOIN team_members as m ON m."UserID"=$1
		JOIN teams as t ON t."ID"=m."TeamID"
		WHERE p."TeamID"=t."ID"
	`, tknData["user_id"])
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	team := models.TeamMember{}

	err = db.DBMap.SelectOne(&team, `SELECT "TeamID" FROM team_members WHERE "UserID"=$1`, tknData["user_id"])
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	for res.Next() {
		err = res.Scan(&count)
		if err != nil {
			myerrors.ServerError(w, r, err)
			return
		}
	}

	if count >= db.GetMaxProjects() {
		myerrors.Respond(w, &myerrors.MySimpleError{
			Code:    http.StatusForbidden,
			Message: "project_limit_reached",
			Req:     r,
		})
		return
	}

	project.Title = postProject.Title
	postProject.Title = ""

	err = db.DBMap.SelectOne(&postProject, `SELECT * FROM workbench_projects WHERE Lower("Title")=Lower($1)`, project.Title)

	if err != sql.ErrNoRows && strings.ToLower(postProject.Title) == strings.ToLower(project.Title) {
		myerrors.Respond(w, &myerrors.MySimpleError{
			Code:    http.StatusConflict,
			Message: "project_exists",
			Req:     r,
		})
		return
	} else if err != sql.ErrNoRows && err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	var newID uint64
	err = db.Instance.QueryRow(`INSERT INTO workbench_projects ("Title", "TeamID") VALUES ($1, $2) RETURNING "ID"`, project.Title, team.TeamID).Scan(&newID)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	project.ID = newID

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(project.PrepareMarshal())

	return
}
