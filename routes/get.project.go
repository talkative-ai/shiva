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

// GetProject router.Route
/* Path: "/project/{id}"
 * Method: "GET"
 * Responds with a models.AumProject
 */
var GetProject = &router.Route{
	Path:       "/workbench/v1/project/{id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}",
	Method:     "GET",
	Handler:    http.HandlerFunc(getProjectHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT},
}

func getProjectHandler(w http.ResponseWriter, r *http.Request) {

	// Parse the project ID from the URL
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

	// Get the project data from the database
	project := &models.AumProject{}
	err = db.DBMap.SelectOne(project, `SELECT * FROM workbench_projects WHERE "ID"=$1`, id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Parse the token data
	token, err := utilities.ParseJTWClaims(w.Header().Get("x-token"))
	tknData := token["data"].(map[string]interface{})

	// Verify access to the project
	member := &models.TeamMember{}
	err = db.DBMap.SelectOne(member, `
		SELECT t."Role"
		FROM workbench_projects AS p
		JOIN team_members AS t
		ON t."TeamID"=p."TeamID" AND t."UserID"=$1
		WHERE p."ID"=$2
	`, tknData["user_id"], id)
	if member.Role != 1 || err != nil {
		// The user has no access
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Get all of the zones
	zones, err := db.DBMap.Select(models.AumZone{}, `SELECT * FROM workbench_zones WHERE "ProjectID"=$1`, id)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	// Attach the zones to the project model. This is for the frontend
	project.Zones = []models.AumZone{}
	for _, zone := range zones {
		project.Zones = append(project.Zones, *zone.(*models.AumZone))
	}

	// Get all of the actors
	actors, err := db.DBMap.Select(models.AumActor{}, `SELECT * FROM workbench_actors WHERE "ProjectID"=$1`, id)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	// Attach the actors to the project model.
	project.Actors = []models.AumActor{}
	for _, actor := range actors {
		project.Actors = append(project.Actors, *actor.(*models.AumActor))
	}

	// Now get all the zone/actor relations
	// We do this because an actor can exist in multiple zones.
	zoneActors, err := db.DBMap.Select(models.AumZoneActor{}, `
		SELECT DISTINCT za."ZoneID", za."ActorID"
		FROM workbench_zones as z
		JOIN workbench_zones_actors as za
		ON za."ZoneID" = z."ID"
		WHERE z."ProjectID"=$1`, id)

	// Attach as above
	project.ZoneActors = []models.AumZoneActor{}
	for _, za := range zoneActors {
		project.ZoneActors = append(project.ZoneActors, *za.(*models.AumZoneActor))
	}

	// Return project data
	json.NewEncoder(w).Encode(project.PrepareMarshal())
}
