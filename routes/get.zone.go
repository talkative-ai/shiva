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

// GetZone router.Route
/* Path: "/zone/{id}"
 * Method: "GET"
 * Responds with models.Zone
 */
var GetZone = &router.Route{
	Path:       "/workbench/v1/zone/{id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}",
	Method:     "GET",
	Handler:    http.HandlerFunc(getZoneHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT},
}

func getZoneHandler(w http.ResponseWriter, r *http.Request) {

	// Parse the Zone ID from the URL
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

	// Fetch the zone data
	zone := &models.Zone{}
	err = db.DBMap.SelectOne(zone, `SELECT * FROM workbench_zones WHERE "ID"=$1`, id)
	if err != nil {
		log.Printf("Zone %+v params %+v", *zone, urlparams)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Parse the JWT data
	token, err := utilities.ParseJTWClaims(w.Header().Get("x-token"))
	tknData := token["data"].(map[string]interface{})

	// Validate access to the project
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

	// Fetch all triggers related to the zone
	// TODO: This could probably be put into a single query
	var triggers []interface{}
	triggers, err = db.DBMap.Select(models.Trigger{}, `
		SELECT * FROM workbench_triggers t
		WHERE t."ZoneID"=$1
	`, zone.ID)

	// Map the triggers to the Zone
	zone.Triggers = map[models.TriggerType]models.Trigger{}
	for _, trigger := range triggers {
		zone.Triggers[trigger.(*models.Trigger).TriggerType] = *trigger.(*models.Trigger)
	}

	// Return zone data
	json.NewEncoder(w).Encode(zone)
}
