package routes

import (
	"encoding/json"
	"net/http"

	"github.com/artificial-universe-maker/go-utilities/models"
	"github.com/artificial-universe-maker/go-utilities/myerrors"
	"github.com/artificial-universe-maker/shiva/router"

	"github.com/artificial-universe-maker/shiva/prehandle"
)

// PatchProject router.Route
// Path: "/user/register",
// Method: "PATCH",
// Accepts models.TokenValidate
// Responds with status of success or failure
var PatchProjects = &router.Route{
	Path:       "/v1/projects",
	Method:     "PATCH",
	Handler:    http.HandlerFunc(patchProjectsHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT, prehandle.RequireBody(65535)},
}

func patchProjectsHandler(w http.ResponseWriter, r *http.Request) {

	project := new(models.AumProject)
	user := new(models.User)

	err := json.Unmarshal([]byte(r.Header.Get("X-Body")), project)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	err = json.Unmarshal([]byte(r.Header.Get("X-User")), user)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	// project.OwnerID = user.Sub

	// generatedIDs := map[string]int64{}

	// tx := db.ShivaDB.MustBegin()

	// for _, zone := range project.Zones {
	// 	if zone.Created != nil {
	// 		res, err := tx.NamedExec("INSERT INTO zones (title) VALUES (:title)", zone)
	// 		if err != nil {
	// 			myerrors.ServerError(w, r, err)
	// 			return
	// 		}
	// 		newID, err := res.LastInsertId()
	// 		if err != nil {
	// 			myerrors.ServerError(w, r, err)
	// 			return
	// 		}
	// 		tx.MustExec("INSERT INTO project_zones VALUES ($1, $2)", project.ID, newID)
	// 		generatedIDs[*zone.Created] = newID
	// 		continue
	// 	}
	// }

	// err = tx.Commit()
	// if err != nil {
	// 	myerrors.ServerError(w, r, err)
	// 	return
	// }

	// resp, err := json.Marshal(generatedIDs)
	// if err != nil {
	// 	myerrors.ServerError(w, r, err)
	// 	return
	// }

	// fmt.Fprintln(w, string(resp))
}
