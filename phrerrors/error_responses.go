package phrerrors

import (
	"encoding/json"
	"fmt"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

type IPHRError interface {
	Parse() (int, string)
}

type PHRSimpleError struct {
	Code    int
	Message interface{}
	Log     string        `json:"-"`
	Req     *http.Request `json:"-"`
}

func (simple *PHRSimpleError) Parse() (int, string) {
	ctx := appengine.NewContext(simple.Req)
	if simple.Log != "" {
		log.Debugf(ctx, simple.Log)
	}

	encoded, err := json.Marshal(simple)
	if err != nil {
		log.Errorf(ctx, fmt.Sprintln("Problem encoding error", err))
	}

	return simple.Code, string(encoded)
}

func Respond(w http.ResponseWriter, err IPHRError) {
	code, msg := err.Parse()
	w.WriteHeader(code)
	fmt.Fprintln(w, msg)
}

func ServerError(w http.ResponseWriter, rq *http.Request, err error) {
	Respond(w, &PHRSimpleError{
		Code: http.StatusInternalServerError,
		Log:  err.Error(),
		Req:  rq,
	})
}
