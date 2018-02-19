package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	utilities "github.com/talkative-ai/core"
	"github.com/talkative-ai/core/db"
	"github.com/talkative-ai/core/models"
	"github.com/talkative-ai/core/myerrors"
	"github.com/talkative-ai/core/router"
	uuid "github.com/talkative-ai/go.uuid"

	"github.com/talkative-ai/core/prehandle"
)

// PatchActor router.Route
/* Path: "/actor/{id}"
 * Method: "PATCH"
 * Accepts a models.Actor, including nested []models.Dialog and []models.DialogRelation.
 * Responds with a map of generated IDs.
 * 		One place that generated IDs comes into play is when multiple objects are created
 *		on the frontend and related to each other in some way.
 *		By enabling this without interfacing with the backend, the workbench can work offline as well
 *		and save/sync when it gets to an internet connection.
 */
var PatchActor = &router.Route{
	Path:       "/workbench/v1/actor/{id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}",
	Method:     "PATCH",
	Handler:    http.HandlerFunc(putActorHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT, prehandle.RequireBody(65535)},
}

func putActorHandler(w http.ResponseWriter, r *http.Request) {

	// Parse the actor ID from the URL
	urlparams := mux.Vars(r)
	actorID, err := uuid.FromString(urlparams["id"])
	if err != nil {
		myerrors.Respond(w, &myerrors.MySimpleError{
			Code:    http.StatusBadRequest,
			Message: "bad_id",
			Req:     r,
		})
		return
	}

	// Fetch the actor ProjectID and simultaneously check if it exists
	actor := &models.Actor{}
	err = db.DBMap.SelectOne(actor, `SELECT "ProjectID" FROM workbench_actors WHERE "ID"=$1`, actorID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	//  We're storing the ProjectID now to make sure the unmarshal won't overwrite it with a bogus one
	projectID := actor.ProjectID

	// Unmarshal the frontend payload into the actor model
	err = json.Unmarshal([]byte(r.Header.Get("X-Body")), actor)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	// Replace the IDs to ensure they're not tampered with by the frontend payload
	actor.ID = actorID
	actor.ProjectID = projectID

	// Parse the JWT
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
	`, tknData["user_id"], actor.ProjectID)
	if member.Role != 1 || err != nil {
		if err != nil {
			myerrors.ServerError(w, r, err)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Open up a database transaction. Either the whole actor updates or nothing at all
	tx := db.Instance.MustBegin()

	// TODO:
	// Way more field validations here
	// Probably generalize validations across models
	db.DBMap.Update(actor)

	// The frontend can create multiple new entities at once,
	// for example multiple new dialogs. Those entities have IDs that are not UUIDs
	// When we store them in the database, we want to tell the frontend which
	// temporary frontend-generated IDs map to the new IDs
	generatedIDs := map[string]uuid.UUID{}

	// This API endpoint also supports patching actor dialogs
	// so we're going to start looping the dialogs
	// and perform any necessary operations.
	for _, dialog := range actor.Dialogs {

		// PatchAction informs us what kind of CRUD operation this is.
		// Although since we'll never "read" then the possible options are create/update/delete
		if dialog.PatchAction == nil {
			// If there's no patch action here we can just skip
			continue
		}

		switch *dialog.PatchAction {
		// We're creating a new dialog
		// This means that the dialog model should definitely have a CreateID
		// If not, something is going wrong
		// TODO: Return an error if no CreateID
		case models.PatchActionCreate:
			// The frontend just nests dialogs within the actor model out of convenience
			// So here we're transforming it back into the shape the database expects
			dialog.ActorID = actorID
			if dialog.CreateID != nil {
				var newID uuid.UUID
				// Prepare these values for the SQL query
				dEntryInput, err := dialog.EntryInput.Value()
				if err != nil {
					myerrors.ServerError(w, r, err)
					return
				}
				dAlwaysExec, err := dialog.AlwaysExec.Value()
				if err != nil {
					myerrors.ServerError(w, r, err)
					return
				}
				dStatements, err := dialog.Statements.Value()
				if err != nil {
					myerrors.ServerError(w, r, err)
					return
				}

				// Create the new dialog
				err = tx.QueryRow(`INSERT INTO 
				workbench_dialog_nodes ("ActorID", "EntryInput", "AlwaysExec", "Statements", "IsRoot", "UnknownHandler")
				VALUES ($1, $2, $3, $4, $5, $6) RETURNING "ID"`, dialog.ActorID, dEntryInput, dAlwaysExec, dStatements, dialog.IsRoot, dialog.UnknownHandler).Scan(&newID)
				if err != nil {
					myerrors.ServerError(w, r, err)
					return
				}

				// Map the frontend-generated temporary ID to the newly generated permanent UUID
				generatedIDs[*dialog.CreateID] = newID
				w.WriteHeader(http.StatusCreated)
				continue
			}
		case models.PatchActionDelete:

			// If we're deleting the actor, we need to delete all connected nodes
			// TODO: Delete from zone relations, and delete actor model itself.
			tx.Exec(`DELETE FROM
				workbench_dialog_nodes_relations
				WHERE "ParentNodeID"=$1 OR "ChildNodeID"=$1`, dialog.ID)
			tx.Exec(`DELETE FROM
				workbench_dialog_nodes
				WHERE "ID"=$1`, dialog.ID)

		case models.PatchActionUpdate:
			// Here we're updating an existing dialog
			dEntryInput, err := dialog.EntryInput.Value()
			if err != nil {
				myerrors.ServerError(w, r, err)
				return
			}
			dAlwaysExec, err := dialog.AlwaysExec.Value()
			if err != nil {
				myerrors.ServerError(w, r, err)
				return
			}
			dStatements, err := dialog.Statements.Value()
			if err != nil {
				myerrors.ServerError(w, r, err)
				return
			}

			// Updating the dialogue in the database.
			tx.Exec(`
				UPDATE workbench_dialog_nodes
				SET "EntryInput" = $1, "AlwaysExec" = $2, "Statements" = $3, "IsRoot" = $4, "UnknownHandler" = $5
				WHERE "ID"=$6 AND "ActorID"=$7
				`, dEntryInput, dAlwaysExec, dStatements, dialog.IsRoot, dialog.UnknownHandler, dialog.ID, actorID)
		}
	}

	for _, relation := range actor.DialogRelations {
		// Same as with the dialogues, but now with the DialogRelations.
		if relation.PatchAction == nil {
			continue
		}

		// If the ChildNodeID has a CreateID
		// then this means that the ChildNode was created on the frontend
		// So we go ahead and set the UUID which would've been generated in the
		// earlier for loop range actor.Dialogs
		if relation.ChildNodeID.CreateID != nil {
			relation.ChildNodeID.UUID = generatedIDs[*relation.ChildNodeID.CreateID]
		}

		// As above, if the ParentNodeID has a CreateID
		// then it means the ParentNode was created on the frontend
		if relation.ParentNodeID.CreateID != nil {
			relation.ParentNodeID.UUID = generatedIDs[*relation.ParentNodeID.CreateID]
		}

		switch *relation.PatchAction {
		case models.PatchActionCreate:
			// Creating the relation
			tx.Exec(`INSERT INTO
				workbench_dialog_nodes_relations ("ParentNodeID", "ChildNodeID")
				VALUES ($1, $2)`, relation.ParentNodeID, relation.ChildNodeID)
		case models.PatchActionDelete:
			// Deleting the relation
			tx.Exec(`DELETE FROM
				workbench_dialog_nodes_relations
				WHERE "ParentNodeID"=$1 AND "ChildNodeID"=$2`, relation.ParentNodeID, relation.ChildNodeID)
		}
	}

	// Finally commit the transaction
	err = tx.Commit()
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	// Return the map of generatedIDs to the frontend
	// so the temporary IDs can be replaced with the new permanent ones
	json.NewEncoder(w).Encode(generatedIDs)
}
