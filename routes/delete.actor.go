package routes

import (
	"encoding/json"
	"log"
	"net/http"

	utilities "github.com/talkative-ai/core"
	"github.com/talkative-ai/core/db"
	"github.com/talkative-ai/core/models"
	"github.com/talkative-ai/core/myerrors"
	"github.com/talkative-ai/core/prehandle"
	"github.com/talkative-ai/core/router"
	uuid "github.com/talkative-ai/go.uuid"

	"github.com/gorilla/mux"
)

// DeleteActor router.Route
// Path: "/actor/{id}"
// Method: "DELETE"
// Responds with the actor data
var DeleteActor = &router.Route{
	Path:       "/workbench/v1/actor/{id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}",
	Method:     "DELETE",
	Handler:    http.HandlerFunc(deleteActorHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT},
}

func deleteActorHandler(w http.ResponseWriter, r *http.Request) {

	urlparams := mux.Vars(r)

	id, err := uuid.FromString(urlparams["id"])
	if err != nil {
		myerrors.Respond(w, &myerrors.MySimpleError{
			Code:    http.StatusBadRequest,
			Message: "bad_id",
			Req:     r,
		})
		return
	}
	// Validate actor access
	actor := &models.AumActor{}
	err = db.DBMap.SelectOne(actor, `SELECT * FROM workbench_actors WHERE "ID"=$1`, id)
	if err != nil {
		log.Printf("Actor %+v params %+v", *actor, urlparams)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	token, err := utilities.ParseJTWClaims(w.Header().Get("x-token"))
	tknData := token["data"].(map[string]interface{})

	member := &models.TeamMember{}
	err = db.DBMap.SelectOne(member, `
		SELECT t."Role"
		FROM workbench_projects AS p
		JOIN team_members AS t
		ON t."TeamID"=p."TeamID" AND t."UserID"=$1
		WHERE p."ID"=$2
	`, tknData["user_id"], actor.ProjectID)
	if member.Role != 1 || err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	_, err = db.DBMap.Delete(*actor)
	if err != nil {
		myerrors.ServerError(w, r, err)
	}

	_, err = db.DBMap.Query(`DELETE FROM workbench_dialog_nodes as dn WHERE db."ActorID"=$1`, actor.ID)
	if err != nil {
		myerrors.ServerError(w, r, err)
	}

	// Return actor data
	json.NewEncoder(w).Encode(actor)
}
