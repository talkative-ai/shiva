package routes

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/talkative-ai/core/db"
	"github.com/talkative-ai/core/models"
	"github.com/talkative-ai/core/myerrors"
	"github.com/talkative-ai/core/prehandle"
	"github.com/talkative-ai/core/router"
	auth "google.golang.org/api/oauth2/v2"
)

// PostAuthGoogle router.Route
/* Path: "/workbench/v1/auth/google"
 * Method: "GET"
 * Validates a Google OAuth 2 token.
 * Responds with a hashmap containing an intercom.io token
 *		and models.User model
 */
var PostAuthGoogle = &router.Route{
	Path:       "/workbench/v1/auth/google",
	Method:     "POST",
	Handler:    http.HandlerFunc(postAuthGoogleHandler),
	Prehandler: []prehandle.Prehandler{prehandle.RequireBody(65535)},
}

type postAuthGooglePayload struct {
	Token      string
	GivenName  string
	FamilyName string
}

func postAuthGoogleHandler(w http.ResponseWriter, r *http.Request) {

	// Create a new Google Auth service
	authService, err := auth.New(http.DefaultClient)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	var tokenInfo *auth.Tokeninfo
	userInfo := postAuthGooglePayload{}
	err = json.Unmarshal([]byte(r.Header.Get("X-Body")), &userInfo)
	if err != nil {
		// If there's an error and it's not that there are no rows
		// Then there's something wonky going on and we'll just return it here.
		myerrors.ServerError(w, r, err)
		return
	}

	if os.Getenv("DEVELOPMENT_ENVIRONMENT") == "TESTING" {
		tokenInfo = &auth.Tokeninfo{
			Email:         "wyatt+test@talkative.ai",
			VerifiedEmail: true,
		}
	} else {
		// Validate the token info on Google's servers
		tokenInfo, err = authService.Tokeninfo().IdToken(userInfo.Token).Do()
		if err != nil {
			// If there's a problem, just assume it's unauthorized.
			myerrors.Respond(w, &myerrors.MySimpleError{
				Code: http.StatusUnauthorized,
				Req:  r,
			})
			return
		}
	}

	if !tokenInfo.VerifiedEmail {
		// If the email hasn't yet been verified, don't allow them access.
		myerrors.Respond(w, &myerrors.MySimpleError{
			Code:    http.StatusForbidden,
			Req:     r,
			Message: "verify_email",
		})
		return
	}

	// Preparing to check if the user has ever signed into the workbench
	newUser := false

	// Check to see if the user exists
	user := &models.User{}
	err = db.DBMap.SelectOne(user, "SELECT * FROM users WHERE \"Email\"=$1", tokenInfo.Email)
	if err != nil && err != sql.ErrNoRows {
		// If there's an error and it's not that there are no rows
		// Then there's something wonky going on and we'll just return it here.
		myerrors.ServerError(w, r, err)
		return
	}

	// Otherwise the user is a new user (does not yet exist in our database)
	if err == sql.ErrNoRows {
		newUser = true
		// Store their details.
		user.Email = tokenInfo.Email
		user.GivenName = userInfo.GivenName
		user.FamilyName = userInfo.FamilyName

		// If the name contains bogus characters, or is empty, or is huge, then don't accept it
		// We want to make sure people are using normal names. This isn't Reddit.
		// TODO: Support other languages and accented characters here.
		if match, err := regexp.MatchString(`[^\w\s\.]|\d`, user.FamilyName); match == true ||
			err != nil ||
			user.FamilyName == "" ||
			len(user.FamilyName) > 50 {
			fmt.Printf("Invalid family name: %v, %v\n", err, user.FamilyName)
			myerrors.Respond(w, &myerrors.MySimpleError{
				Code:    http.StatusBadRequest,
				Req:     r,
				Message: "invalid_family_name",
			})
			return
		}

		// Same as above, but with the last name.
		// TODO: Same as above
		if match, err := regexp.MatchString(`[^\w\s\.]|\d`, user.GivenName); match == true ||
			err != nil ||
			user.GivenName == "" ||
			len(user.GivenName) > 50 {
			fmt.Printf("Invalid given name: %v, %v\n", err, user.GivenName)
			myerrors.Respond(w, &myerrors.MySimpleError{
				Code:    http.StatusBadRequest,
				Req:     r,
				Message: "invalid_given_name",
			})
			return
		}

		// Create and save the user
		if err := db.CreateAndSaveUser(user); err != nil {
			myerrors.ServerError(w, r, err)
			return
		}
	}

	// Generate their token
	// TODO: Create a nice token struct
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Minute * 60 * 24 * 30).Unix(),
		"data": map[string]interface{}{
			"user_id": user.ID,
		},
	})

	// Sign the token and stringify
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_KEY")))
	if err != nil {
		log.Println("Error", err)
		return
	}

	// Attach the JWT to the x-token header
	w.Header().Set("x-token", tokenString)

	if newUser {
		// If it's a new user, then let them know
		w.WriteHeader(http.StatusCreated)
	} else {
		// Otherwise just say OK
		w.WriteHeader(http.StatusOK)
	}

	// Generate an Intercom token so that we can verify the user identity on intercom.io
	mac := hmac.New(sha256.New, []byte(os.Getenv("INTERCOM_SECRET")))
	mac.Write([]byte(user.Email))

	// Pass all the data to the user
	json.NewEncoder(w).Encode(map[string]interface{}{
		"IntercomHMAC": hex.EncodeToString(mac.Sum(nil)),
		"User":         user,
	})

}
