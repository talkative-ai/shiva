package routes

import (
	"fmt"
	"net/http"

	"github.com/artificial-universe-maker/shiva/prehandle"
	"github.com/artificial-universe-maker/shiva/router"
	auth "google.golang.org/api/oauth2/v2"
)

// PostAuthgoogle router.Route
// Path: "/auth/google",
// Method: "POST",
// Accepts models.UserRegister
// Responds with status of success or failure
var PostAuthgoogle = &router.Route{
	Path:       "/v1/auth/google",
	Method:     "GET",
	Handler:    http.HandlerFunc(postAuthGoogleHandler),
	Prehandler: []prehandle.Prehandler{},
}

func postAuthGoogleHandler(w http.ResponseWriter, r *http.Request) {

	authService, err := auth.New(http.DefaultClient)
	if err != nil {
		fmt.Println(err)
		return
	}
	tokenInfo, err := authService.Tokeninfo().IdToken(r.FormValue("token")).Do()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%+v", tokenInfo)

}
