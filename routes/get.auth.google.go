package routes

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/artificial-universe-maker/go-utilities/db"
	"github.com/artificial-universe-maker/go-utilities/models"
	"github.com/artificial-universe-maker/go-utilities/myerrors"
	"github.com/artificial-universe-maker/shiva/prehandle"
	"github.com/artificial-universe-maker/shiva/router"
	jwt "github.com/dgrijalva/jwt-go"
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
	newUser := false

	// Check to see if the user exists
	user := &models.User{}
	err = db.DBMap.SelectOne(user, "SELECT * FROM users WHERE \"Email\"=$1", tokenInfo.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			// User does not exist. Create and initialize base team
			newUser = true
			user.Email = tokenInfo.Email
			user.GivenName = r.FormValue("gn")
			user.FamilyName = r.FormValue("fn")
			if match, err := regexp.MatchString(`\W`, user.FamilyName); match == true ||
				err != nil ||
				user.FamilyName == "" ||
				len(user.FamilyName) > 50 {
				myerrors.Respond(w, &myerrors.MySimpleError{
					Code:    http.StatusBadRequest,
					Req:     r,
					Message: "invalid_family_name",
				})
				return
			}
			if match, err := regexp.MatchString(`\W`, user.GivenName); match == true ||
				err != nil ||
				user.GivenName == "" ||
				len(user.GivenName) > 50 {
				myerrors.Respond(w, &myerrors.MySimpleError{
					Code:    http.StatusBadRequest,
					Req:     r,
					Message: "invalid_given_name",
				})
				return
			}
			if err := db.CreateAndSaveUser(user); err != nil {
				myerrors.ServerError(w, r, err)
				return
			}
		} else {
			myerrors.ServerError(w, r, err)
			return
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Minute * 60 * 24 * 30).Unix(),
		"id":  user.ID,
	})

	tokenString, err := token.SignedString([]byte("secret"))
	if err != nil {
		log.Println("Error", err)
		return
	}

	w.Header().Set("x-token", tokenString)

	if newUser {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(*user)

}
