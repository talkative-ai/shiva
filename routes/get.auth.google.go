package routes

import (
	"fmt"
	"net/http"

	"github.com/artificial-universe-maker/go-utilities/db"
	"github.com/artificial-universe-maker/go-utilities/models"
	"github.com/artificial-universe-maker/go-utilities/myerrors"
	"github.com/artificial-universe-maker/shiva/prehandle"
	"github.com/artificial-universe-maker/shiva/router"
	auth "google.golang.org/api/oauth2/v2"
)

// PostAuthGoogle router.Route
// Path: "/v1/auth/google",
// Method: "GET",
// Validates a Google OAuth 2 token.
// Responds with status of success or failure
var PostAuthGoogle = &router.Route{
	Path:       "/v1/auth/google",
	Method:     "GET",
	Handler:    http.HandlerFunc(postAuthGoogleHandler),
	Prehandler: []prehandle.Prehandler{},
}

func postAuthGoogleHandler(w http.ResponseWriter, r *http.Request) {

	authService, err := auth.New(http.DefaultClient)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	tokenInfo, err := authService.Tokeninfo().IdToken(r.FormValue("token")).Do()
	if err != nil {
		myerrors.Respond(w, &myerrors.MySimpleError{
			Code: http.StatusUnauthorized,
			Log:  err.Error(),
			Req:  r,
		})
		return
	}

	if !tokenInfo.VerifiedEmail {
		myerrors.Respond(w, &myerrors.MySimpleError{
			Code:    http.StatusForbidden,
			Req:     r,
			Message: "verify_email",
		})
		return
	}

	err = db.InitializeDB()

	// Check to see if the user exists
	user := &models.User{}
	_, err = db.DBMap.Select(user, "SELECT * FROM users WHERE email=$1", tokenInfo.Email)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	newUser := false

	fmt.Printf("%+v", user)

	// User does not exist. Create and initialize base team
	if user.ID == 0 {
		newUser = true
		user.Email = tokenInfo.Email
		err := db.CreateAndSaveUser(user)
		if err != nil {
			myerrors.ServerError(w, r, err)
			return
		}
	}

	if newUser {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}

}
