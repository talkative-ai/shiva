package routes

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	utilities "github.com/artificial-universe-maker/go-utilities"
	"github.com/artificial-universe-maker/go-utilities/db"
	"github.com/artificial-universe-maker/go-utilities/models"
	"github.com/artificial-universe-maker/go-utilities/myerrors"
	"github.com/artificial-universe-maker/go-utilities/prehandle"
	"github.com/artificial-universe-maker/go-utilities/router"

	"github.com/gorilla/mux"
)

// GetActor router.Route
// Path: "/actor/{id}",
// Method: "GET",
// Accepts models.TokenValidate
// Responds with the actor data
var GetActor = &router.Route{
	Path:       "/v1/actor/{id:[0-9]+}",
	Method:     "GET",
	Handler:    http.HandlerFunc(getActorHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT},
}

func getActorHandler(w http.ResponseWriter, r *http.Request) {

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
	log.Println(tknData)

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

	dialogNodes, err := db.DBMap.Select(models.AumDialogNode{}, `
		SELECT *
		FROM workbench_dialog_nodes as dn
		WHERE dn."ActorID"=$1`, id)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	actor.Dialogs = []models.AumDialogNode{}
	for _, dn := range dialogNodes {
		actor.Dialogs = append(actor.Dialogs, *dn.(*models.AumDialogNode))
	}

	dialogRelations, err := db.DBMap.Select(models.AumDialogRelation{}, `
		SELECT DISTINCT dr.*
		FROM workbench_dialog_nodes as dn
		JOIN workbench_dialog_nodes_relations as dr
		ON dr."ParentNodeID" = dn."ID"
		OR dr."ChildNodeID" = dn."ID"
		WHERE dn."ActorID"=$1`, id)

	actor.DialogRelations = []models.AumDialogRelation{}
	for _, dr := range dialogRelations {
		actor.DialogRelations = append(actor.DialogRelations, *dr.(*models.AumDialogRelation))
	}

	// Return actor data
	json.NewEncoder(w).Encode(actor)
}