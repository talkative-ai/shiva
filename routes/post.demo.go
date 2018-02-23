package routes

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	utilities "github.com/talkative-ai/core"
	"github.com/talkative-ai/core/db"
	"github.com/talkative-ai/core/models"
	"github.com/talkative-ai/core/myerrors"
	uuid "github.com/talkative-ai/go.uuid"

	"github.com/talkative-ai/core/prehandle"
	"github.com/talkative-ai/core/router"
)

// PostDemo router.Route
/* Path: "/publish",
 * Method: "POST",
 * Accepts models.UserRegister
 * Responds with status of success or failure
 */
var PostDemo = &router.Route{
	Path:       "/workbench/v1/demo/{id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}",
	Method:     "POST",
	Handler:    http.HandlerFunc(postDemoHandler),
	Prehandler: []prehandle.Prehandler{prehandle.JWT},
}

func postDemoHandler(w http.ResponseWriter, r *http.Request) {

	urlparams := mux.Vars(r)

	token, _ := utilities.ParseJTWClaims(w.Header().Get("x-token"))
	tknData := token["data"].(map[string]interface{})

	id, err := uuid.FromString(urlparams["id"])
	if err != nil {
		myerrors.Respond(w, &myerrors.MySimpleError{
			Code:    http.StatusBadRequest,
			Message: "bad_id",
			Req:     r,
		})
		return
	}

	member := &models.TeamMember{}
	err = db.DBMap.SelectOne(member, `
		SELECT t."Role"
		FROM workbench_projects AS p
		JOIN team_members AS t
		ON t."TeamID"=p."TeamID" AND t."UserID"=$1
		WHERE p."ID"=$2
	`, tknData["user_id"], id)
	if member.Role != 1 || err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	rq, err := http.NewRequest("POST", fmt.Sprintf("http://lakshmi:8080/v1/publish/%v?demo=true", urlparams["id"]), nil)
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
