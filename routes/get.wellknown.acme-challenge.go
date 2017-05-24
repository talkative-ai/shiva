package routes

import (
	"io"
	"net/http"

	"github.com/warent/stdapi/router"

	"github.com/warent/stdapi/prehandle"
)

// GetWellknownAcmeChallenge router.Route
// Method: "GET"
// SSL Verification method -- Must be updated every 3 months
// Last update: April 23, 2017
var GetWellknownAcmeChallenge = &router.Route{
	Path:       "/.well-known/acme-challenge/_wt6Wp8oG8DIGCHDCB0JKRl0UFbxLUwydnGT4WAyuf0",
	Method:     "GET",
	Handler:    http.HandlerFunc(GetWellknownAcmeChallengeHandler),
	Prehandler: []prehandle.Prehandler{},
}

func GetWellknownAcmeChallengeHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "_wt6Wp8oG8DIGCHDCB0JKRl0UFbxLUwydnGT4WAyuf0.eAqIP9l15IZFgWuMXfi_L1lwiD-53l7pz_q1ENaRX_Q")
}
