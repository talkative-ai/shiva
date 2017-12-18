package routes

import (
	"encoding/json"
	"net/http"

	utilities "github.com/artificial-universe-maker/core"
	"github.com/artificial-universe-maker/core/db"
	"github.com/artificial-universe-maker/core/models"
	"github.com/artificial-universe-maker/core/myerrors"
	"github.com/artificial-universe-maker/core/router"
	uuid "github.com/artificial-universe-maker/go.uuid"
	"github.com/gorilla/mux"

	"github.com/artificial-universe-maker/core/prehandle"
)

// PatchActor router.Route
// Path: "/actor/{id}",
// Method: "PATCH",
// Accepts models.TokenValidate
// Responds with status of success or failure
var PatchActor = &router.Route{
	Path:       "/workbench/v1/actor/{id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}",
	Method:     "PATCH",
	Handler:    http.HandlerFunc(putActorHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT, prehandle.RequireBody(65535)},
}

func putActorHandler(w http.ResponseWriter, r *http.Request) {

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
	actor := &models.AumActor{}
	err = db.DBMap.SelectOne(actor, `SELECT "ProjectID" FROM workbench_actors WHERE "ID"=$1`, actorID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	projectID := actor.ProjectID

	err = json.Unmarshal([]byte(r.Header.Get("X-Body")), actor)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	actor.ID = actorID
	actor.ProjectID = projectID

	token, err := utilities.ParseJTWClaims(w.Header().Get("x-token"))
	tknData := token["data"].(map[string]interface{})

	// Validate user has actor access
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

	tx := db.Instance.MustBegin()

	// TODO:
	// Way more field validations here
	// Probably generalize validations across models
	db.DBMap.Update(actor)

	generatedIDs := map[string]uuid.UUID{}

	for _, dialog := range actor.Dialogs {
		if dialog.PatchAction == nil {
			continue
		}

		switch *dialog.PatchAction {
		case models.PatchActionCreate:
			dialog.ActorID = actorID
			// Creating a new dialog
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

				err = tx.QueryRow(`INSERT INTO 
				workbench_dialog_nodes ("ActorID", "EntryInput", "AlwaysExec", "Statements", "IsRoot")
				VALUES ($1, $2, $3, $4, $5) RETURNING "ID"`, dialog.ActorID, dEntryInput, dAlwaysExec, dStatements, dialog.IsRoot).Scan(&newID)
				if err != nil {
					myerrors.ServerError(w, r, err)
					return
				}
				generatedIDs[*dialog.CreateID] = newID
				continue
			}
		case models.PatchActionDelete:
			tx.Exec(`DELETE FROM
				workbench_dialog_nodes_relations
				WHERE "ParentNodeID"=$1 OR "ChildNodeID"=$1`, dialog.ID)
			tx.Exec(`DELETE FROM
				workbench_dialog_nodes
				WHERE "ID"=$1`, dialog.ID)
		case models.PatchActionUpdate:
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

			// TODO: Generalize this with reflection
			tx.Exec(`
				UPDATE workbench_dialog_nodes
				SET "EntryInput" = $1, "AlwaysExec" = $2, "Statements" = $3, "IsRoot" = $4, "UnknownHandler" = $5
				WHERE "ID"=$6 AND "ActorID"=$7
				`, dEntryInput, dAlwaysExec, dStatements, dialog.IsRoot, dialog.UnknownHandler, dialog.ID, actorID)
		}
	}

	for _, relation := range actor.DialogRelations {
		if relation.PatchAction == nil {
			continue
		}

		if relation.ChildNodeID.CreateID != nil {
			relation.ChildNodeID.UUID = generatedIDs[*relation.ChildNodeID.CreateID]
		}
		if relation.ParentNodeID.CreateID != nil {
			relation.ParentNodeID.UUID = generatedIDs[*relation.ParentNodeID.CreateID]
		}

		switch *relation.PatchAction {
		case models.PatchActionCreate:
			tx.Exec(`INSERT INTO
				workbench_dialog_nodes_relations ("ParentNodeID", "ChildNodeID")
				VALUES ($1, $2)`, relation.ParentNodeID, relation.ChildNodeID)
		case models.PatchActionDelete:
			tx.Exec(`DELETE FROM
				workbench_dialog_nodes_relations
				WHERE "ParentNodeID"=$1 AND "ChildNodeID"=$2`, relation.ParentNodeID, relation.ChildNodeID)
		}
	}

	err = tx.Commit()
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	json.NewEncoder(w).Encode(generatedIDs)
}
