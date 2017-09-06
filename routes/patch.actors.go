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

// PatchActor router.Route
// Path: "/actor/{id}",
// Method: "PATCH",
// Accepts models.TokenValidate
// Responds with status of success or failure
var PatchActors = &router.Route{
	Path:       "/v1/actor/{id:[0-9]+}",
	Method:     "PATCH",
	Handler:    http.HandlerFunc(patchActorsHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT, prehandle.RequireBody(65535)},
}

func patchActorsHandler(w http.ResponseWriter, r *http.Request) {

	urlparams := mux.Vars(r)

	actorID, err := strconv.ParseInt(urlparams["id"], 10, 64)
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

	// Validate user has actor access
	member := &models.TeamMember{}
	err = db.DBMap.SelectOne(member, `
			SELECT t."Role"
			FROM workbench_actors AS p
			JOIN team_members AS t
			ON t."TeamID"=p."TeamID" AND t."UserID"=$1
			WHERE p."ID"=$2
		`, tknData["user_id"], actorID)
	if member.Role != 1 || err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	actor := new(models.AumActor)

	err = json.Unmarshal([]byte(r.Header.Get("x-body")), actor)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	tx := db.Instance.MustBegin()

	generatedIDs := map[int]uint64{}

	// for _, dialog := range actor.Dialogs {
	// 	if dialog.CreateID != nil {
	// 		var newID uint64
	// 		lblock := models.RawLBlock{}
	// 		json.Unmarshal()
	// 		err = tx.QueryRow(`INSERT INTO workbench_dialogs ("ActorID", "Title") VALUES ($1, $2) RETURNING "ID"`, actorID, dialog.Title).Scan(&newID)
	// 		if err != nil {
	// 			myerrors.ServerError(w, r, err)
	// 			return
	// 		}
	// 		generatedIDs[*dialog.CreateID] = newID
	// 		continue
	// 	} else {

	// 		if err != nil {
	// 			myerrors.ServerError(w, r, err)
	// 			return
	// 		}
	// 	}
	// }

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
