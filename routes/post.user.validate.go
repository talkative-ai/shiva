package routes

import (
	"fmt"
	"net/http"

	"github.com/phrhero/calcifer/models"
	"github.com/phrhero/stdapi/aeproviders"
	"github.com/phrhero/stdapi/phrerrors"
	"github.com/phrhero/stdapi/router"

	"encoding/json"

	"github.com/phrhero/stdapi/prehandle"
	"google.golang.org/appengine"
)

// PostUserValidate router.Route
// Path: "/user/validate",
// Method: "POST",
// Accepts models.UserValidate
// Responds with status of success or failure
var PostUserValidate = &router.Route{
	Path:       "/v1/user/validate",
	Method:     "POST",
	Handler:    http.HandlerFunc(postUserValidateHandler),
	Prehandler: []prehandle.Prehandler{prehandle.RequireBody(1024)},
}

func postUserValidateHandler(w http.ResponseWriter, r *http.Request) {

	ctx := appengine.NewContext(r)

	cache, err := aeproviders.AEConnectRedis(ctx)
	if err != nil {
		phrerrors.ServerError(w, r, err)
		return
	}

	defer func() {
		if cache != nil {
			cache.Close()
		}
	}()

	userParams := &models.StdParams{Cache: cache, W: w, R: r}
	var user models.User
	err = json.Unmarshal([]byte(r.Header.Get("X-Body")), &user)
	if err != nil {
		phrerrors.ServerError(w, r, err)
		return
	}

	isValid, err := user.Validate(userParams)
	if err != nil {
		phrerrors.ServerError(w, r, err)
		return
	}

	var encoded []byte

	if !isValid {
		encoded, _ = json.Marshal(map[string]string{
			"status": "E_NOT_VALID",
		})
		fmt.Fprintln(w, string(encoded))
		return
	}

	isVerified, err := user.HasAccountFlag(userParams, models.UserAccountEmailVerified)
	if err != nil {
		phrerrors.ServerError(w, r, err)
		return
	}
	if !isVerified {
		encoded, _ = json.Marshal(map[string]string{
			"status": "VERIFICATION_SENT",
		})
		fmt.Fprintln(w, string(encoded))
		return
	}

	encoded, _ = json.Marshal(map[string]string{
		"status": "VALID",
	})
	fmt.Fprintln(w, string(encoded))
	return

}
