package routes

import (
	"fmt"
	"net/http"

	"github.com/talkative-ai/core/models"
	"github.com/talkative-ai/core/myerrors"
	"github.com/talkative-ai/core/prehandle"
	"github.com/talkative-ai/core/redis"
	"github.com/talkative-ai/core/router"
	uuid "github.com/talkative-ai/go.uuid"

	"github.com/gorilla/mux"
)

// GetDemoStatus router.Route
/* Path: "/project/{id}/metadata"
 * Method: "GET"
 * Responds with an integer representing the demo publish status
 */
var GetDemoStatus = &router.Route{
	Path:       "/workbench/v1/demo/{id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}/status",
	Method:     "GET",
	Handler:    http.HandlerFunc(getDemoStatusHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.JWT},
}

func getDemoStatusHandler(w http.ResponseWriter, r *http.Request) {

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

	status := redis.Instance.Get(fmt.Sprintf("%v:%v", models.KeynavProjectMetadataStatic(fmt.Sprintf("demo:%v", id.String())), "status")).Val()

	fmt.Fprint(w, status)
}
