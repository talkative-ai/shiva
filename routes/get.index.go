package routes

import (
	"fmt"
	"net/http"

	"github.com/warent/phrhero-backend/router"

	"github.com/warent/phrhero-backend/prehandle"
)

// GetIndex router.Route
// Path: "/user/register",
// Method: "POST",
// Accepts models.UserRegister
// Responds with status of success or failure
var GetIndex = &router.Route{
	Path:       "/",
	Method:     "GET",
	Handler:    http.HandlerFunc(getIndexHandler),
	Prehandler: []prehandle.Prehandler{},
}

func getIndexHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "Want to help us build this api and more? Email us: dev-jobs@phrhero.com")

}
