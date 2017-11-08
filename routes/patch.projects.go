package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	utilities "github.com/artificial-universe-maker/go-utilities"
	"github.com/artificial-universe-maker/go-utilities/db"
	"github.com/artificial-universe-maker/go-utilities/models"
	"github.com/artificial-universe-maker/go-utilities/myerrors"
	"github.com/artificial-universe-maker/go-utilities/router"
	"github.com/gorilla/mux"

	"github.com/artificial-universe-maker/go-utilities/prehandle"
)

// PatchProject router.Route
// Path: "/user/register",
// Method: "PATCH",
// Accepts models.TokenValidate
/**
	PatchProject enables the creation of new entities within a project.
	In order to update existing entities, use a Put{Entity} endpoint.
**/
var PatchProject = &router.Route{
	Path:       "/workbench/v1/project/{id:[0-9]+}",
	Method:     "PATCH",
	Handler:    http.HandlerFunc(patchProjectsHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT, prehandle.RequireBody(65535)},
}

func patchProjectsHandler(w http.ResponseWriter, r *http.Request) {

	urlparams := mux.Vars(r)

	projectID, err := strconv.ParseUint(urlparams["id"], 10, 64)
	if err != nil {
		myerrors.Respond(w, &myerrors.MySimpleError{
			Code:    http.StatusBadRequest,
			Message: "bad_id",
			Req:     r,
		})
		return
	}

	token, err := utilities.ParseJTWClaims(w.Header().Get("x-token"))
	tknData := token["data"].(map[string]interface{})

	// Validate user has project access
	member := &models.TeamMember{}
	err = db.DBMap.SelectOne(member, `
			SELECT t."Role"
			FROM workbench_projects AS p
			JOIN team_members AS t
			ON t."TeamID"=p."TeamID" AND t."UserID"=$1
			WHERE p."ID"=$2
		`, tknData["user_id"], projectID)
	if member.Role != 1 || err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	project := new(models.AumProject)

	err = json.Unmarshal([]byte(r.Header.Get("x-body")), project)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	tx := db.Instance.MustBegin()

	generatedIDs := map[string]uint64{}

	for _, zone := range project.Zones {
		if zone.CreateID != nil {
			var newID uint64
			err = tx.QueryRow(`INSERT INTO workbench_zones ("ProjectID", "Title") VALUES ($1, $2) RETURNING "ID"`, projectID, zone.Title).Scan(&newID)
			if err != nil {
				myerrors.ServerError(w, r, err)
				return
			}
			w.WriteHeader(http.StatusCreated)
			generatedIDs[*zone.CreateID] = newID
			continue
		}
	}

	for _, actor := range project.Actors {
		if actor.CreateID != nil {
			var newID uint64
			err = tx.QueryRow(`INSERT INTO workbench_actors ("ProjectID", "Title") VALUES ($1, $2) RETURNING "ID"`, projectID, actor.Title).Scan(&newID)
			if err != nil {
				myerrors.ServerError(w, r, err)
				return
			}
			w.WriteHeader(http.StatusCreated)
			generatedIDs[*actor.CreateID] = newID
			actor.ID = newID
		}
	}

	for _, za := range project.ZoneActors {
		if za.PatchAction == nil {
			continue
		}

		switch v := za.ActorID.(type) {
		// If the ActorID is a string, then this is a CreateID
		case string:
			za.ActorID = generatedIDs[v]
		}
		switch v := za.ZoneID.(type) {
		// If the ZoneID is a string, then this is a CreateID
		case string:
			za.ZoneID = generatedIDs[v]
		}

		switch *za.PatchAction {
		case models.PatchActionCreate:
			tx.Exec(`INSERT INTO
				workbench_zones_actors ("ZoneID", "ActorID")
				VALUES ($1, $2)`, za.ZoneID, za.ActorID)
		case models.PatchActionDelete:
			tx.Exec(`DELETE FROM workbench_zones_actors WHERE "ZoneID"=$1 AND "ActorID"=$2`, za.ZoneID, za.ActorID)
		}
	}

	err = tx.Commit()
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	resp, err := json.Marshal(generatedIDs)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	fmt.Fprintln(w, string(resp))
}
