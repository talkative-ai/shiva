package routes

import (
	"net/http"

	"github.com/warent/phrhero-backend/router"

	"github.com/warent/phrhero-backend/prehandle"
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

	return
}
