package routes

import (
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
		return
	}

	created, err := user.SetAccountFlag(userParams, models.UserAccountExists)
	if err != nil {
		log.Errorf(ctx, err.Error())
		return
	}

	if created < 1 {
		log.Debugf(ctx, "%d", created)
		return
	}

	log.Debugf(ctx, "Here")

	if err = user.EncryptPassword(); err != nil {
		log.Errorf(ctx, "post.user.register.go handler: %s", err.Error())
		return
	}

	user.RegisterAccount(userParams)
	if stat, err := user.SendVerificationEmail(userParams); err != nil {
		log.Errorf(ctx, err.Error())
	} else {
		log.Debugf(ctx, "%d", stat)
	}

}
