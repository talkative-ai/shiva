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
	"github.com/artificial-universe-maker/core/redis"
	"github.com/artificial-universe-maker/core/router"
	uuid "github.com/artificial-universe-maker/go.uuid"

	"github.com/gorilla/mux"
)

// GetProjectMetadata router.Route
/* Path: "/project/{id}/metadata"
 * Method: "GET"
 * Responds with an ad hoc models.ProjectMetadata containing the publish status and time.
 *		Optionally contains the rejection reason of the latest submission if applicable.
 */
var GetProjectMetadata = &router.Route{
	Path:       "/workbench/v1/project/{id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}/metadata",
	Method:     "GET",
	Handler:    http.HandlerFunc(getProjectMetadataHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT},
}

func getProjectMetadataHandler(w http.ResponseWriter, r *http.Request) {

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

	// Fetch the project data from the database
	project := &models.AumProject{}
	err = db.DBMap.SelectOne(project, `SELECT * FROM workbench_projects WHERE "ID"=$1`, id)
	if err != nil {
		log.Printf("Project %+v params %+v", *project, urlparams)
		w.WriteHeader(http.StatusNotFound)
		return
	}

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
	`, tknData["user_id"], id)
	if member.Role != 1 || err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// We store publishing information into Redis.
	// The reason why is that the data may be accessed by the assistant
	// and all Google Home data is saved in Redis for the purpose of speed.
	// The first is the publish status
	status := redis.Instance.Get(fmt.Sprintf("%v:%v", models.KeynavProjectMetadataStatic(id.String()), "status")).Val()
	// pubtime is when it was published
	pubtime := redis.Instance.Get(fmt.Sprintf("%v:%v", models.KeynavProjectMetadataStatic(id.String()), "pubtime")).Val()

	// Parse integers from the returned string
	// TODO: Could this just be .Int64() from the Redis client rather than .Val()?
	statusNum, err := strconv.ParseInt(status, 10, 8)
	pubtimeNum, err := strconv.ParseInt(pubtime, 10, 64)
	pubtimeParsed := time.Unix(0, pubtimeNum)

	metadata := &models.ProjectMetadata{
		Status:      models.PublishStatus(statusNum),
		PublishTime: pubtimeParsed,
	}

	// If the project was denied the last time it was submitted by the user
	// then we fetch the reason why it was rejected to show on the frontend.
	if metadata.Status == models.PublishStatusDenied {
		metadata.Review = &models.ProjectReviewPublic{}
		err = db.DBMap.SelectOne(metadata.Review, `
			SELECT
				"BadTitle", "MajorProblems", "MinorProblems", "ProblemWith", "Dialogues"
			FROM project_review_results
			WHERE "ProjectID"=$1
			ORDER BY "ReviewedAt" DESC
			LIMIT 1
		`, id.String())
		if err != nil {
			myerrors.ServerError(w, r, err)
			return
		}
	}

	// Return the project metadata
	json.NewEncoder(w).Encode(metadata)
}
