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
	Method:     "DELETE",
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
