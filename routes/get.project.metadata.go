package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	utilities "github.com/artificial-universe-maker/core"
	"github.com/artificial-universe-maker/core/db"
	"github.com/artificial-universe-maker/core/models"
	"github.com/artificial-universe-maker/core/myerrors"
	"github.com/artificial-universe-maker/core/prehandle"
	"github.com/artificial-universe-maker/core/providers"
	"github.com/artificial-universe-maker/core/router"
	uuid "github.com/artificial-universe-maker/go.uuid"

	"github.com/gorilla/mux"
)

// GetProject router.Route
// Path: "/project/{id}",
// Method: "GET",
// Accepts models.TokenValidate
// Responds with the project data
var GetProjectMetadata = &router.Route{
	Path:       "/workbench/v1/project/{id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}/metadata",
	Method:     "GET",
	Handler:    http.HandlerFunc(getProjectMetadataHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT},
}

func getProjectMetadataHandler(w http.ResponseWriter, r *http.Request) {

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

	redis, err := providers.ConnectRedis()
	if err != nil {
		myerrors.Respond(w, &myerrors.MySimpleError{
			Code: http.StatusInternalServerError,
			Log:  err.Error(),
			Req:  r,
		})
		return
	}
	defer redis.Close()

	type ProjectMetadata struct {
		Status      models.PublishStatus
		PublishTime time.Time
	}

	status := redis.Get(fmt.Sprintf("%v:%v", models.KeynavProjectMetadataStatic(id.String()), "status")).Val()
	pubtime := redis.Get(fmt.Sprintf("%v:%v", models.KeynavProjectMetadataStatic(id.String()), "pubtime")).Val()

	statusNum, err := strconv.ParseInt(status, 10, 8)
	pubtimeNum, err := strconv.ParseInt(pubtime, 10, 64)
	pubtimeParsed := time.Unix(0, pubtimeNum)

	metadata := &ProjectMetadata{
		Status:      models.PublishStatus(statusNum),
		PublishTime: pubtimeParsed,
	}

	json.NewEncoder(w).Encode(metadata)
}
