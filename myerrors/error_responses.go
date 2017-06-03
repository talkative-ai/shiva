package myerrors

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/op/go-logging"
)

type IMyError interface {
	Parse() (int, string)
}

type MySimpleError struct {
	Code    int           `json:"code"`
	Message interface{}   `json:"message"`
	Log     string        `json:"-"`
	Req     *http.Request `json:"-"`
}

func (simple *MySimpleError) Parse() (int, string) {

	var log = logging.MustGetLogger("Log")

	if simple.Log != "" {
		log.Debugf(simple.Log)
	}

	encoded, err := json.Marshal(simple)
	if err != nil {
		log.Errorf(fmt.Sprintln("Problem encoding error", err))
	}

	return simple.Code, string(encoded)
}

func Respond(w http.ResponseWriter, err IMyError) {
	code, msg := err.Parse()
	w.WriteHeader(code)
	fmt.Fprintln(w, msg)
}

func ServerError(w http.ResponseWriter, rq *http.Request, err error) {
	Respond(w, &MySimpleError{
		Code: http.StatusInternalServerError,
		Log:  err.Error(),
		Req:  rq,
	})
}
