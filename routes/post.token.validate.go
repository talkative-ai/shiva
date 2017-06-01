package routes

import (
	"encoding/json"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/urlfetch"

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

	ctx := appengine.NewContext(r)

	client := urlfetch.Client(ctx)
	info, err := GoogleIdTokenVerifier.Verify(token.Token, "895662102905-6369ghd23tqhvrv9t26lfjmobj3hgmfn.apps.googleusercontent.com", client)

	if err != nil {
		myerrors.Respond(w, &myerrors.MySimpleError{
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
			Code: http.StatusBadRequest,
			Log:  err.Error(),
			Message: map[string]string{
				"message": "VERIFY_EMAIL",
			},
		})
		return
	}

	k := datastore.NewIncompleteKey(ctx, "User", nil)

	type user struct {
		Email string
	}

	u := &user{
		info.Email,
	}

	if _, err := datastore.Put(ctx, k, u); err != nil {
		myerrors.ServerError(w, r, err)
		return
	}
}
