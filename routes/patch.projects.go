package routes

import (
	"encoding/json"
	"net/http"

	"github.com/jmoiron/sqlx"

	"github.com/gorilla/mux"
	utilities "github.com/talkative-ai/core"
	"github.com/talkative-ai/core/db"
	"github.com/talkative-ai/core/models"
	"github.com/talkative-ai/core/myerrors"
	"github.com/talkative-ai/core/router"
	uuid "github.com/talkative-ai/go.uuid"

	"github.com/talkative-ai/core/prehandle"
)

// PatchProject router.Route
/* Path: "/project/{id}",
 * Method: "PATCH",
 * Accepts models.TokenValidate
 * Accepts a models.Actor, including nested []models.Dialog and []models.DialogRelation.
 * Responds with a map of generated IDs.
 * 		One place that generated IDs comes into play is when multiple objects are created
 *		on the frontend and related to each other in some way.
 *		By enabling this without interfacing with the backend, the workbench can work offline as well
 *		and save/sync when it gets to an internet connection.
 */
var PatchProject = &router.Route{
	Path:       "/workbench/v1/project/{id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}",
	Method:     "PATCH",
	Handler:    http.HandlerFunc(patchProjectsHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT, prehandle.RequireBody(65535)},
}

func patchProjectsHandler(w http.ResponseWriter, r *http.Request) {

	// Parse the project ID from the URL
	urlparams := mux.Vars(r)
	projectID, err := uuid.FromString(urlparams["id"])
	if err != nil {
		myerrors.Respond(w, &myerrors.MySimpleError{
			Code:    http.StatusBadRequest,
			Message: "bad_id",
			Req:     r,
		})
		return
	}

	// Parse the JWT data
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

	// Fetching the start zone of the project
	// This will be used later because if there's no StartZoneID
	// then we set it to default to the first zone created
	proj := &models.Project{}
	db.DBMap.SelectOne(proj, `
		SELECT "StartZoneID"
		FROM workbench_projects
		WHERE "ID"=$1
		`, projectID)

	project := new(models.Project)

	// Parse the frontend payload into the project model
	err = json.Unmarshal([]byte(r.Header.Get("x-body")), project)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	// Reset the projectID to the one we validated
	// to prevent tampering
	project.ID = projectID
	project.StartZoneID = proj.StartZoneID

	// Create a transaction
	tx := db.Instance.MustBegin()

	// Prepare the generatedIDs map
	// Entities on the frontend are created with temporary IDs.
	// We use this map to communicate the entities' new permanent IDs.
	generatedIDs := map[string]uuid.UUID{}

	// First we update the project zones
	for _, zone := range project.Zones {

		// If the CreateID has a value, then the frontend has generated a temp
		// ID for this zone, it's a new zone, and we need to insert it into the database
		if zone.CreateID != nil {

			if len(zone.Title) == 0 || len(zone.Title) > 255 {
				myerrors.Respond(w, &myerrors.MySimpleError{
					Code:    http.StatusBadRequest,
					Message: "bad_zone_title",
					Req:     r,
				})
				return
			}

			var newID uuid.UUID
			err = tx.QueryRow(`
				INSERT INTO workbench_zones
				("ProjectID", "Title")
				VALUES ($1, $2)
				RETURNING "ID"`, projectID, zone.Title).Scan(&newID)
			if err != nil {
				myerrors.ServerError(w, r, err)
				return
			}
			w.WriteHeader(http.StatusCreated)

			// Map the new ID to the old frontend temp ID
			generatedIDs[*zone.CreateID] = newID

			// If the StartZoneID for the project isn't set, default it to this zone.
			// This is good UX when a user creates their first zone to the project.
			if !project.StartZoneID.Valid || project.StartZoneID.UUID == uuid.Nil {
				project.StartZoneID.UUID = newID
				project.StartZoneID.Valid = true
				// We're updating the project start zone here
				// TODO: Implement a better method to validate whether this is necessary
				updateProjectStartZone(tx, project.StartZoneID.UUID, project.ID)
			}
		}

		// Updating the triggers within the zones
		for t, trigger := range zone.Triggers {

			if trigger.PatchAction == nil {
				return
			}

			if *trigger.PatchAction == models.PatchActionDelete {
				// Deleting the trigger
				tx.Exec(`
					DELETE FROM workbench_triggers
					WHERE "ProjectID"=$1
					AND "TriggerType"=$2
					AND "ZoneID"=$3`, project.ID, trigger.TriggerType, trigger.ZoneID)
				continue
			}

			if *trigger.PatchAction == models.PatchActionCreate {
				// Creating a new trigger
				trigger.TriggerType = t
				if trigger.ZoneID.CreateID != nil {
					// The trigger was added to a zone that was created on the frontend but never stored on the backend
					// therefore we need to update the ZoneID with the new permanent ID
					trigger.ZoneID.UUID = generatedIDs[*trigger.ZoneID.CreateID]
				}
				// Prepare the non-standard model values to be stored in the database
				execPrepared, err := trigger.AlwaysExec.Value()
				if err != nil {
					myerrors.ServerError(w, r, err)
					return
				}

				// Store
				_, err = tx.Exec(`
					INSERT INTO workbench_triggers
						("ProjectID", "ZoneID", "TriggerType", "AlwaysExec")
					SELECT $1, $2, $3, $4
					WHERE EXISTS (
						SELECT "ID" FROM workbench_zones
						WHERE "ProjectID" = $1
						AND "ID" = $2
					)`, project.ID, zone.ID, trigger.TriggerType, execPrepared)
				if err != nil {
					myerrors.ServerError(w, r, err)
					return
				}
				w.WriteHeader(http.StatusCreated)
				continue
			}

			if *trigger.PatchAction == models.PatchActionUpdate {
				// Updating an existing trigger
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

	// Updating the actors
	for _, actor := range project.Actors {
		// The only updating that will happen here is creating new actors.
		// Updating existing actors actually happens in the patch.actor endpoint
		if actor.CreateID == nil {
			myerrors.Respond(w, &myerrors.MySimpleError{
				Code:    http.StatusBadRequest,
				Message: "missing_create_id",
				Req:     r,
			})
			return
		}

		if len(actor.Title) == 0 || len(actor.Title) > 255 {
			myerrors.Respond(w, &myerrors.MySimpleError{
				Code:    http.StatusBadRequest,
				Message: "bad_actor_title",
				Req:     r,
			})
			return
		}

		var newID uuid.UUID
		// New actors have a CreateID set. Otherwise they already exist
		// Store the new actor.
		err = tx.QueryRow(`
				INSERT INTO workbench_actors
					("ProjectID", "Title")
				VALUES ($1, $2) RETURNING "ID"`, projectID, actor.Title).Scan(&newID)
		if err != nil {
			myerrors.ServerError(w, r, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		// Mapping the old ID and the new ID
		generatedIDs[*actor.CreateID] = newID
		actor.ID = newID
	}

	// Updating the actors relationships to the zones
	// In other words, adding or removing actors to zones
	for _, za := range project.ZoneActors {
		if za.PatchAction == nil {
			continue
		}

		// If the zone didn't previously exist, then we need to update
		// the zoneID here to the permanent one
		if za.ZoneID.CreateID != nil {
			za.ZoneID.UUID = generatedIDs[*za.ZoneID.CreateID]
		}

		// Same as above, but updating the ActorID to the permanent one if it was only just created
		if za.ActorID.CreateID != nil {
			za.ActorID.UUID = generatedIDs[*za.ActorID.CreateID]
		}

		switch *za.PatchAction {
		case models.PatchActionCreate:
			// The actor was added to the zone
			// Create the new relation into the DB
			tx.Exec(`INSERT INTO
				workbench_zones_actors ("ZoneID", "ActorID")
				SELECT $1, $2
				WHERE NOT EXISTS (
					SELECT "ZoneID" FROM workbench_zones_actors
					WHERE "ZoneID" = $1
					AND "ActorID" = $2
				)`, za.ZoneID, za.ActorID)
		case models.PatchActionDelete:
			// The actor was removed from the zone
			tx.Exec(`
				DELETE FROM workbench_zones_actors
				WHERE "ZoneID"=$1
				AND "ActorID"=$2
				AND EXISTS (
					SELECT wz."ProjectID" FROM workbench_zones wz
					INNER JOIN workbench_actors wa
					ON wa."ProjectID"=wz."ProjectID"
					AND wa."ID"=$2
					WHERE wz."ID"=$1
					AND wz."ProjectID"=$3
				)`, za.ZoneID, za.ActorID, projectID)
		}
	}

	// We're updating the project start zone here
	// TODO: Implement a better method to validate whether this is necessary
	updateProjectStartZone(tx, project.StartZoneID.UUID, projectID)

	// Updating the project tags
	// TODO: Attach patch actions to this to avoid unnecessary queries?
	// TODO: Input validation here
	if project.Tags != nil {
		tags, _ := project.Tags.Value()
		_, err = tx.Exec(`
			UPDATE workbench_projects
			SET "Tags"=$1
			WHERE "ID"=$2
			`, tags, projectID)
	}

	// Updating the project categories
	// TODO: Input validation here. Also attach a patch action
	if project.Category != nil {
		_, err = tx.Exec(`
			UPDATE workbench_projects
			SET "Category"=$1
			WHERE "ID"=$2
			`, project.Category, projectID)
	}

	// Commit the update
	err = tx.Commit()
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	// Send the new IDs list to the frontend to replace the old temp ones
	json.NewEncoder(w).Encode(generatedIDs)
}

// This is a helper function for updating the project start zone
func updateProjectStartZone(context *sqlx.Tx, StartZoneID, projectID uuid.UUID) bool {

	// If the zone is Nil for some reason, then abort
	if StartZoneID == uuid.Nil {
		return false
	}

	// First validate that the zone exists in the project
	var count int
	context.QueryRow(`
		SELECT COUNT(*) FROM workbench_zones
		WHERE "ID"=$1 AND "ProjectID"=$2`,
		StartZoneID, projectID).Scan(&count)
	if count <= 0 {
		// If not, then abort
		return false
	}

	// Update the value on the project
	context.Exec(`
			UPDATE workbench_projects
			SET "StartZoneID"=$1
			WHERE "ID"=$2
			`, StartZoneID, projectID)
	return true
}
