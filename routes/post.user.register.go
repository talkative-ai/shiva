package routes

import (
	"net/http"

	"github.com/warent/phrhero-backend/errors"
	"github.com/warent/phrhero-backend/models"
	"github.com/warent/phrhero-backend/router"

	"github.com/warent/phrhero-backend/prehandle"
	"github.com/warent/phrhero-backend/providers"
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
		errors.ServerError(w, r, err)
		return
	}

	defer func() {
		if cache != nil {
			cache.Close()
		}
	}()

	userParams := &models.StdParams{Cache: cache, W: w, R: r}

	user := &models.User{
		Email:     "",
		FirstName: "",
		LastName:  "",
	}

	accStatus, err := user.GetAccountStatus(userParams)
	if err != nil {
		return
	}

	if accStatus != models.USER_ACCOUNT_DNE {
		// Account exists
		return
	}

	isNewAccount, err := user.SetAccountStatus(userParams, models.USER_ACCOUNT_CREATING)
	if err != nil {
		return
	}

	if !isNewAccount {
		// Duplicate accounts are being created simultaneously. Abort
		return
	}

}
