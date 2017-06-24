package routes

import (
	"io"
	"net/http"

	"github.com/artificial-universe-maker/shiva/router"

	"github.com/artificial-universe-maker/shiva/prehandle"
)

// GetWellknownAcmeChallenge router.Route
// Method: "GET"
// SSL Verification method -- Must be updated every 3 months
// Last update: June 3, 2017
var GetWellknownAcmeChallenge = &router.Route{
	Path:       "/.well-known/acme-challenge/HZU7RNYx-6vCMJgyYCSsb9cWq7tUytrXSgDmwBk-9os",
	Method:     "GET",
	Handler:    http.HandlerFunc(GetWellknownAcmeChallengeHandler),
	Prehandler: []prehandle.Prehandler{},
}

func GetWellknownAcmeChallengeHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "HZU7RNYx-6vCMJgyYCSsb9cWq7tUytrXSgDmwBk-9os.J4na4u-fBxUI5TEx_YZvpq8yUwNKMSLya4IyAU96Q68")
}
