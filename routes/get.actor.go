package routes

import (
	"encoding/json"
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

// GetActor router.Route
// Path: "/actor/{id}"
// Method: "GET"
// Responds with the corresponding models.Actor data
var GetActor = &router.Route{
	Path:       "/workbench/v1/actor/{id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}",
	Method:     "GET",
	Handler:    http.HandlerFunc(getActorHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT},
}

func getActorHandler(w http.ResponseWriter, r *http.Request) {

	urlparams := mux.Vars(r)

	// Get the Actor ID from the URL
	id, err := uuid.FromString(urlparams["id"])
	if err != nil {
		// If the ID isn't a valid UUID, return a bad request error
		myerrors.Respond(w, &myerrors.MySimpleError{
			Code:    http.StatusBadRequest,
			Message: "bad_id",
			Req:     r,
		})
		return
	}

	// Getting the entity data from the database
	actor := &models.Actor{}
	err = db.DBMap.SelectOne(actor, `SELECT * FROM workbench_actors WHERE "ID"=$1`, id)
	if err != nil {
		// Unable to find the actor
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Parse the current JWT token data
	token, err := utilities.ParseJTWClaims(w.Header().Get("x-token"))
	tknData := token["data"].(map[string]interface{})

	// Begin validating user access to the project
	member := &models.TeamMember{}
	err = db.DBMap.SelectOne(member, `
		SELECT t."Role"
		FROM workbench_projects AS p
		JOIN team_members AS t
		ON t."TeamID"=p."TeamID" AND t."UserID"=$1
		WHERE p."ID"=$2
	`, tknData["user_id"], actor.ProjectID)
	if member.Role != 1 || err != nil {
		// If the user is not of the correct role or there's an error
		// return with unauthorized.
		// TODO: Place roles in a const
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Fetching all of the dialog nodes to a dialog item
	// These are stored in the database as their own entities,
	// but we combine it for frontend convenience
	dialogNodes, err := db.DBMap.Select(models.DialogNode{}, `
		SELECT *
		FROM workbench_dialog_nodes as dn
		WHERE dn."ActorID"=$1`, id)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	// Map the dalogs to the actor
	actor.Dialogs = []models.DialogNode{}
	for _, dn := range dialogNodes {
		actor.Dialogs = append(actor.Dialogs, *dn.(*models.DialogNode))
	}

	// Same with the dialog nodes, only this time we're fetching their relations
	dialogRelations, err := db.DBMap.Select(models.DialogRelation{}, `
		SELECT DISTINCT dr.*
		FROM workbench_dialog_nodes as dn
		JOIN workbench_dialog_nodes_relations as dr
		ON dr."ParentNodeID" = dn."ID"
		OR dr."ChildNodeID" = dn."ID"
		WHERE dn."ActorID"=$1`, id)

	// Map to the actor
	actor.DialogRelations = []models.DialogRelation{}
	for _, dr := range dialogRelations {
		actor.DialogRelations = append(actor.DialogRelations, *dr.(*models.DialogRelation))
	}

	// Return actor data
	json.NewEncoder(w).Encode(actor)
}
