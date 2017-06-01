package routes

import (
	"encoding/json"
	"log"
	"net/http"

	apiai "github.com/warent/apiai-go"
	"github.com/warent/stdapi/prehandle"
	"github.com/warent/stdapi/router"
)

// PostAction router.Route
// Path: "/user/register",
// Method: "POST",
// Accepts models.UserRegister
// Responds with status of success or failure
var PostAction = &router.Route{
	Path:       "/v1/action",
	Method:     "POST",
	Handler:    http.HandlerFunc(postActionHandler),
	Prehandler: []prehandle.Prehandler{},
}

func postActionHandler(w http.ResponseWriter, r *http.Request) {

	input := &apiai.QueryResponse{}

	w.Header().Add("content-type", "application/json")

	var resp map[string]string

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(input)
	if err != nil {
		log.Println("Error", err)
		return
	}

	if input.Result.Action == "input.welcome" {
		resp = map[string]string{
			"speech":      "Well hello there!",
			"displayText": "Well hello there!",
		}
	} else {
		resp = map[string]string{
			"speech":      "Yes, yes, okay then. Isn't that nice?",
			"displayText": "Yes, yes, okay then. Isn't that nice?",
		}
	}

	json.NewEncoder(w).Encode(resp)
}
