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

// GetProject router.Route
// Path: "/project/{id}",
// Method: "GET",
// Accepts models.TokenValidate
// Responds with the project data
var GetProject = &router.Route{
	Path:       "/v1/project/{id:[0-9]+}",
	Method:     "GET",
	Handler:    http.HandlerFunc(getProjectHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT},
}

func getProjectHandler(w http.ResponseWriter, r *http.Request) {

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

	// Validate project access
	project := &models.AumProject{}
	err = db.DBMap.SelectOne(project, `SELECT * FROM workbench_projects WHERE "ID"=$1`, id)
	if err != nil {
		log.Printf("Project %+v params %+v", *project, urlparams)
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
	`, tknData["user_id"], id)
	if member.Role != 1 || err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	zones, err := db.DBMap.Select(models.AumZone{}, `SELECT * FROM workbench_zones WHERE "ProjectID"=$1`, id)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	project.Zones = []models.AumZone{}
	for _, zone := range zones {
		project.Zones = append(project.Zones, *zone.(*models.AumZone))
	}

	actors, err := db.DBMap.Select(models.AumActor{}, `SELECT * FROM workbench_actors WHERE "ProjectID"=$1`, id)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	project.Actors = []models.AumActor{}
	for _, actor := range actors {
		project.Actors = append(project.Actors, *actor.(*models.AumActor))
	}

	zoneActors, err := db.DBMap.Select(models.AumZoneActor{}, `
		SELECT DISTINCT za."ZoneID", za."ActorID"
		FROM workbench_zones as z
		JOIN workbench_zones_actors as za
		ON za."ZoneID" = z."ID"
		WHERE z."ProjectID"=$1`, id)

	project.ZoneActors = []models.AumZoneActor{}
	for _, za := range zoneActors {
		project.ZoneActors = append(project.ZoneActors, *za.(*models.AumZoneActor))
	}
	// Return project data
	json.NewEncoder(w).Encode(project.PrepareMarshal())
}
