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
	"google.golang.org/appengine/log"
)

// PostUserRegister router.Route
// Path: "/user/register",
// Method: "POST",
// Accepts models.UserRegister
// Responds with status of success or failure
var PostUserRegister = &router.Route{
	Path:       "/v1/user/register",
	Method:     "POST",
	Handler:    http.HandlerFunc(postUserRegisterHandler),
	Prehandler: []prehandle.Prehandler{prehandle.RequireBody(1024)},
}

func postUserRegisterHandler(w http.ResponseWriter, r *http.Request) {

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

	created, err := user.SetAccountFlag(userParams, models.UserAccountExists)
	if err != nil {
		phrerrors.ServerError(w, r, err)
		return
	}

	if created < 1 {
		isVerified, err := user.HasAccountFlag(userParams, models.UserAccountEmailVerified)
		if err != nil {
			phrerrors.ServerError(w, r, err)
			return
		}

		var encoded []byte
		if isVerified {
			encoded, _ = json.Marshal(map[string]string{
				"status": "E_EXISTS",
			})
		} else {
			encoded, _ = json.Marshal(map[string]string{
				"status": "VERIFICATION_SENT",
			})
		}

		fmt.Fprintln(w, string(encoded))
		return
	}

	log.Debugf(ctx, "Here")

	if err = user.EncryptPassword(); err != nil {
		log.Errorf(ctx, "post.user.register.go handler: %s", err.Error())
		return
	}

	stat, err := user.SendVerificationEmail(userParams)
	if err != nil {
		phrerrors.ServerError(w, r, err)
		return
	}

	user.RegisterAccount(userParams)

	encoded, _ := json.Marshal(map[string]string{
		"status": string(stat),
	})
	fmt.Fprintln(w, string(encoded))

}
