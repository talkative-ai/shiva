package routes

import (
	"net/http"

	"phrhero-backend/phrerrors"
	"phrhero-backend/models"
	"phrhero-backend/router"

	"encoding/json"

	"phrhero-backend/prehandle"
	"phrhero-backend/providers"
)

// PostUserRegister router.Route
// Path: "/user/register",
// Method: "POST",
// Accepts models.UserRegister
// Responds with status of success or failure
var PostUserRegister = &router.Route{
	Path:       "/user/register",
	Method:     "POST",
	Handler:    http.HandlerFunc(handler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.RequireBody(1024)},
}

func handler(w http.ResponseWriter, r *http.Request) {

	cache, err := providers.ConnectRedis(r)
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
	json.NewDecoder(r.Body).Decode(&user)

	accStatus, err := user.GetAccountStatus(userParams)
	if err != nil {
		return
	}

	if accStatus&models.USER_ACCOUNT_DNE == 0 {
		// Account exists
		return
	}

	accStatus ^= models.USER_ACCOUNT_DNE
	accStatus |= models.USER_ACCOUNT_CREATING

	isNewAccount, err := user.SetAccountStatus(userParams, accStatus)
	if err != nil {
		return
	}

	if !isNewAccount {
		// Duplicate accounts are being created simultaneously. Abort
		// TODO Handle more elegantly
		return
	}

	user.Save(userParams)
	user.SendVerificationEmail(userParams)

}
