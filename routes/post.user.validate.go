package routes

import (
	"fmt"
	"net/http"

	"github.com/warent/shiva/models"
	"github.com/warent/stdapi/aeproviders"
	"github.com/warent/stdapi/myerrors"
	"github.com/warent/stdapi/router"

	"encoding/json"

	"github.com/warent/stdapi/prehandle"
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
		myerrors.ServerError(w, r, err)
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
		myerrors.ServerError(w, r, err)
		return
	}

	validateStatus, err := user.Validate(userParams)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	var encoded []byte

	if validateStatus == models.NOT_EXIST {
		encoded, _ = json.Marshal(map[string]string{
			"status": "E_NOT_EXIST",
		})
		fmt.Fprintln(w, string(encoded))
		return
	}

	if validateStatus == models.NOT_VALID {
		encoded, _ = json.Marshal(map[string]string{
			"status": "E_NOT_VALID",
		})
		fmt.Fprintln(w, string(encoded))
		return
	}

	isVerified, err := user.HasAccountFlag(userParams, models.UserAccountEmailVerified)
	if err != nil {
		myerrors.ServerError(w, r, err)
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
