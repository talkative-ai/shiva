package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/warent/stdapi/myerrors"
	"github.com/warent/stdapi/router"

	"github.com/warent/stdapi/prehandle"
	"github.com/warent/stdapi/utilities"
)

// PostTokenValidate router.Route
// Path: "/user/register",
// Method: "POST",
// Accepts models.TokenValidate
// Responds with status of success or failure
var PostJWTValidate = &router.Route{
	Path:       "/v1/jwt/validate",
	Method:     "POST",
	Handler:    http.HandlerFunc(postTokenValidateHandler),
	Prehandler: []prehandle.Prehandler{prehandle.RequireBody(5120)},
}

func postJWTValidateHandler(w http.ResponseWriter, r *http.Request) {

	type request struct {
		JWT string
	}

	reqBody := &request{}

	err := json.Unmarshal([]byte(r.Header.Get("X-Body")), reqBody)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	jwttoken, err := utilities.ParseJTWClaims(reqBody.JWT)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	log.Println(jwttoken)

	response, err := json.Marshal(jwttoken["User"])
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	fmt.Fprintln(w, string(response))
}
