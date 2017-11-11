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

// GetZone router.Route
// Path: "/zone/{id}",
// Method: "GET",
// Accepts models.TokenValidate
// Responds with the zone data
var GetZone = &router.Route{
	Path:       "/workbench/v1/zone/{id:[0-9]+}",
	Method:     "GET",
	Handler:    http.HandlerFunc(getZoneHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT},
}

func getZoneHandler(w http.ResponseWriter, r *http.Request) {

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

	// Fetch the zone
	zone := &models.AumZone{}
	err = db.DBMap.SelectOne(zone, `SELECT * FROM workbench_zones WHERE "ID"=$1`, id)
	if err != nil {
		log.Printf("Zone %+v params %+v", *zone, urlparams)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	token, err := utilities.ParseJTWClaims(w.Header().Get("x-token"))
	tknData := token["data"].(map[string]interface{})
	log.Println(tknData)

	// Validate access
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

	var triggers []interface{}

	triggers, err = db.DBMap.Select(models.AumTrigger{}, `
		SELECT * FROM workbench_triggers t
		JOIN workbench_zones_triggers zt
		ON zt."TriggerID"=t."ID"
		AND zt."ZoneID"=$1
	`, zone.ID)

	zone.Triggers = map[models.AumTriggerType]models.AumTrigger{}
	for _, trigger := range triggers {
		zone.Triggers[trigger.(*models.AumTrigger).TriggerType] = *trigger.(*models.AumTrigger)
	}

	// Return zone data
	json.NewEncoder(w).Encode(zone)
}
