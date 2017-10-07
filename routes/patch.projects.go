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
	Path:       "/v1/project/{id:[0-9]+}",
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

	generatedIDs := map[int]uint64{}
	zoneExists := map[uint64]bool{}

	for _, zone := range project.Zones {
		if zone.CreateID != nil {
			var newID uint64
			err = tx.QueryRow(`INSERT INTO workbench_zones ("ProjectID", "Title") VALUES ($1, $2) RETURNING "ID"`, projectID, zone.Title).Scan(&newID)
			if err != nil {
				myerrors.ServerError(w, r, err)
				return
			}
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
			generatedIDs[*actor.CreateID] = newID
			actor.ID = newID
		}
		if actor.ZoneID != nil {
			if _, ok := zoneExists[*actor.ZoneID]; !ok {
				zone := &models.AumZone{}
				err = db.DBMap.SelectOne(zone, `
					SELECT z."ID"
					FROM workbench_zones as z
					WHERE z."ID"=$1 AND z."ProjectID"=$2
				`, actor.ZoneID, projectID)
				if err != nil {
					myerrors.Respond(w, &myerrors.MySimpleError{
						Code: http.StatusUnauthorized,
						Log:  err.Error(),
						Req:  r,
					})
					return
				}
				zoneExists[*actor.ZoneID] = true
			}
			_, err = tx.Exec(`INSERT INTO workbench_zones_actors ("ZoneID", "ActorID") VALUES ($1, $2)`, actor.ZoneID, actor.ID)
			if err != nil {
				myerrors.ServerError(w, r, err)
				return
			}
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
