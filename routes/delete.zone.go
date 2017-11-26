package routes

import (
	"net/http"
	"strconv"

	utilities "github.com/artificial-universe-maker/core"
	"github.com/artificial-universe-maker/core/db"
	"github.com/artificial-universe-maker/core/models"
	"github.com/artificial-universe-maker/core/myerrors"
	"github.com/artificial-universe-maker/core/prehandle"
	"github.com/artificial-universe-maker/core/router"

	"github.com/gorilla/mux"
)

// DeleteZone router.Route
// Path: "/zone/{id}",
// Method: "GET",
// Accepts models.TokenValidate
// Responds with the zone data
var DeleteZone = &router.Route{
	Path:       "/workbench/v1/zone/{id:[0-9]+}",
	Method:     "DELETE",
	Handler:    http.HandlerFunc(deleteZoneHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT},
}

func deleteZoneHandler(w http.ResponseWriter, r *http.Request) {

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

	// Validate zone access
	zone := &models.AumZone{}
	err = db.DBMap.SelectOne(zone, `SELECT * FROM workbench_zones WHERE "ID"=$1`, id)
	if err != nil {
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
	`, tknData["user_id"], zone.ProjectID)
	if member.Role != 1 || err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	_, err = db.DBMap.Delete(*zone)
	if err != nil {
		myerrors.ServerError(w, r, err)
	}
}
