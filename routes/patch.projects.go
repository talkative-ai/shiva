package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	utilities "github.com/artificial-universe-maker/core"
	"github.com/artificial-universe-maker/core/db"
	"github.com/artificial-universe-maker/core/models"
	"github.com/artificial-universe-maker/core/myerrors"
	"github.com/artificial-universe-maker/core/router"
	"github.com/gorilla/mux"

	"github.com/artificial-universe-maker/core/prehandle"
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

	project.ID = projectID

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
		for t, trigger := range zone.Triggers {

			if trigger.PatchAction == nil {
				return
			}

			if *trigger.PatchAction == models.PatchActionDelete {
				tx.Exec(`DELETE FROM workbench_triggers WHERE "ProjectID"=$1 AND "TriggerType"=$2`, project.ID, trigger.TriggerType)
				continue
			}

			if *trigger.PatchAction == models.PatchActionCreate {
				trigger.TriggerType = t
				switch v := trigger.ZoneID.(type) {
				// If the ZoneID is a string, then this is a CreateID
				case string:
					trigger.ZoneID = generatedIDs[v]
				}
				execPrepared, err := trigger.AlwaysExec.Value()
				if err != nil {
					myerrors.ServerError(w, r, err)
					return
				}
				_, err = tx.Exec(`
					INSERT INTO workbench_triggers ("ProjectID", "ZoneID", "TriggerType", "AlwaysExec")
					VALUES ($1, $2, $3, $4)`, project.ID, zone.ID, trigger.TriggerType, execPrepared)
				if err != nil {
					myerrors.ServerError(w, r, err)
					return
				}
				w.WriteHeader(http.StatusCreated)
				continue
			}

			if *trigger.PatchAction == models.PatchActionUpdate {
				execPrepared, err := trigger.AlwaysExec.Value()
				if err != nil {
					myerrors.ServerError(w, r, err)
					return
				}
				_, err = tx.Exec(`
					UPDATE workbench_triggers
					SET "AlwaysExec" = $1
					WHERE "ProjectID" = $2 AND "ZoneID" = $3 AND "TriggerType" = $4`,
					execPrepared, projectID, trigger.ZoneID, t)
				continue
			}
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
			tx.Exec(`DELETE FROM workbench_zones_actors WHERE "ZoneID"=$1 AND "ActorID"=$2`, za.ZoneID, za.ActorID)
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
