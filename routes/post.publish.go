package routes

import (
	"fmt"
	"net/http"

	utilities "github.com/artificial-universe-maker/go-utilities"
	"github.com/artificial-universe-maker/go-utilities/db"
	"github.com/artificial-universe-maker/go-utilities/models"
	"github.com/artificial-universe-maker/go-utilities/myerrors"
	"github.com/gorilla/mux"

	"github.com/artificial-universe-maker/go-utilities/prehandle"
	"github.com/artificial-universe-maker/go-utilities/router"
)

// PostPublish router.Route
// Path: "/publish",
// Method: "POST",
// Accepts models.UserRegister
// Responds with status of success or failure
var PostPublish = &router.Route{
	Path:       "/workbench/v1/publish/{id:[0-9]+}",
	Method:     "POST",
	Handler:    http.HandlerFunc(postPublishHandler),
	Prehandler: []prehandle.Prehandler{prehandle.JWT},
}

func postPublishHandler(w http.ResponseWriter, r *http.Request) {

	urlparams := mux.Vars(r)

	token, _ := utilities.ParseJTWClaims(w.Header().Get("x-token"))
	tknData := token["data"].(map[string]interface{})

	member := &models.TeamMember{}
	err := db.DBMap.SelectOne(member, `
		SELECT t."Role"
		FROM workbench_projects AS p
		JOIN team_members AS t
		ON t."TeamID"=p."TeamID" AND t."UserID"=$1
		WHERE p."ID"=$2
	`, tknData["user_id"], urlparams["id"])
	if member.Role != 1 || err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	rq, err := http.NewRequest("GET", fmt.Sprintf("http://lakshmi:8080/?project-id=%v", urlparams["id"]), nil)
	if err != nil {
		myerrors.ServerError(w, r, err)
	}
	client := http.Client{}
	resp, err := client.Do(rq)
	if err != nil {
		myerrors.ServerError(w, r, err)
	}
	w.WriteHeader(resp.StatusCode)

}
