package routes

import (
	"encoding/json"
	"log"
	"net/http"

	"cloud.google.com/go/datastore"

	"github.com/warent/GoogleIdTokenVerifier"
	"github.com/warent/stdapi/myerrors"
	"github.com/warent/stdapi/router"

	"github.com/warent/stdapi/prehandle"
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
	Prehandler: []prehandle.Prehandler{prehandle.RequireBody(5120)},
}

func postTokenValidateHandler(w http.ResponseWriter, r *http.Request) {

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
			Req:  r,
			Code: http.StatusBadRequest,
			Log:  err.Error(),
			Message: map[string]string{
				"message": "Bad token",
			},
		})
		return
	}

	if !info.EmailVerified {
		myerrors.Respond(w, &myerrors.MySimpleError{
			Req:  r,
			Code: http.StatusBadRequest,
			Message: map[string]string{
				"message": "VERIFY_EMAIL",
			},
		})
		return
	}

	dsClient, err := datastore.NewClient(ctx, "artificial-universe-maker")
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	k := datastore.NameKey("User", info.Sub, nil)

	type user struct {
		Sub     string
		Email   string
		Name    string
		Picture string
	}

	u := &user{
		info.Sub,
		info.Email,
		info.Name,
		info.Picture,
	}

	log.Printf("%v+", u)

	if _, err := dsClient.Put(ctx, k, u); err != nil {
		myerrors.ServerError(w, r, err)
		return
	}
}
