package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"cloud.google.com/go/datastore"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/artificial-universe-maker/GoogleIdTokenVerifier"
	"github.com/artificial-universe-maker/go-utilities/models"
	"github.com/artificial-universe-maker/shiva/myerrors"
	"github.com/artificial-universe-maker/shiva/router"

	"time"

	"github.com/artificial-universe-maker/shiva/prehandle"
)

// PostTokenValidate router.Route
// Path: "/user/register",
// Method: "POST",
// Accepts models.TokenValidate
// Responds with status of success or failure
var PostTokenValidate = &router.Route{
	Path:       "/v1/token/validate",
	Method:     "POST",
	Handler:    http.HandlerFunc(postTokenValidateHandler),
	Prehandler: []prehandle.Prehandler{prehandle.SetJSON, prehandle.RequireBody(5120)},
}

func postTokenValidateHandler(w http.ResponseWriter, r *http.Request) {

	type response struct {
		JWT string
	}

	type Token struct {
		Token    string `json:"token"`
		Provider string `json:"provider"`
	}

	token := &Token{}

	err := json.Unmarshal([]byte(r.Header.Get("X-Body")), token)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	ctx := r.Context()

	info, err := GoogleIdTokenVerifier.Verify(token.Token, "895662102905-6369ghd23tqhvrv9t26lfjmobj3hgmfn.apps.googleusercontent.com", nil)

	if err != nil {
		myerrors.Respond(w, &myerrors.MySimpleError{
			Req:     r,
			Code:    http.StatusBadRequest,
			Log:     err.Error(),
			Message: "TOKEN_BAD",
		})
		return
	}

	if !info.EmailVerified {
		myerrors.Respond(w, &myerrors.MySimpleError{
			Req:     r,
			Code:    http.StatusBadRequest,
			Message: "EMAIL_VERIFY",
		})
		return
	}

	dsClient, err := datastore.NewClient(ctx, "artificial-universe-maker")
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	k := datastore.NameKey("User", info.Sub, nil)

	u := &models.User{
		Sub:     info.Sub,
		Email:   info.Email,
		Name:    info.Name,
		Picture: info.Picture,
	}

	if _, err := dsClient.Put(ctx, k, u); err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	userString, err := json.Marshal(u)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	jwttoken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"Exp":  time.Now().Add(time.Minute * 10).Unix(),
		"User": userString,
	})

	tokenString, err := jwttoken.SignedString([]byte(os.Getenv("JWT_KEY")))
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	resp, err := json.Marshal(&response{
		tokenString,
	})
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	fmt.Fprintln(w, string(resp))
}
