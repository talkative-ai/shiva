package routes

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	utilities "github.com/talkative-ai/core"
	"github.com/talkative-ai/core/db"
	"github.com/talkative-ai/core/models"
	"github.com/talkative-ai/core/myerrors"
	"github.com/talkative-ai/core/prehandle"
	"github.com/talkative-ai/core/router"
	uuid "github.com/talkative-ai/go.uuid"
)

// PostProject router.Route
/* Path: "/project",
 * Method: "POST",
 * Accepts models.TokenValidate
 * Responds with status of success or failure
 */
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

	project := new(models.Project)
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

	// Enforce project limit
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

	// Enforce minimum project title length
	if len(project.Title) < 3 || len(project.Title) > 50 {
		myerrors.Respond(w, &myerrors.MySimpleError{
			Code:    http.StatusBadRequest,
			Message: "bad_title_length",
			Req:     r,
		})
		return
	}

	// Disallow special characters
	match, _ := regexp.MatchString(`[^\w\s!]`, project.Title)
	if match {
		myerrors.Respond(w, &myerrors.MySimpleError{
			Code:    http.StatusBadRequest,
			Message: "bad_title_characters",
			Req:     r,
		})
		return
	}

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

	var newID uuid.UUID
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
