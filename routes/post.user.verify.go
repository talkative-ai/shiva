package routes

import (
	"fmt"
	"net/http"

	"github.com/warent/shiva/models"
	"github.com/warent/stdapi/aeproviders"
	"github.com/warent/stdapi/myerrors"
	"github.com/warent/stdapi/router"

	"encoding/json"

	"github.com/go-redis/redis"
	"github.com/warent/stdapi/prehandle"
	"google.golang.org/appengine"
)

// PostUserVerify router.Route
// Path: "/user/register",
// Method: "POST",
// Accepts models.UserVerify
// Responds with status of success or failure
var PostUserVerify = &router.Route{
	Path:       "/v1/user/verify",
	Method:     "POST",
	Handler:    http.HandlerFunc(postUserVerifyHandler),
	Prehandler: []prehandle.Prehandler{prehandle.RequireBody(1024)},
}

type PostUserVerifyRequest struct {
	Hash string
}

func postUserVerifyHandler(w http.ResponseWriter, r *http.Request) {

	ctx := appengine.NewContext(r)

	cache, err := aeproviders.AEConnectRedis(ctx)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	defer func() {
		if cache != nil {
			cache.Close()
		}
	}()

	var params PostUserVerifyRequest
	err = json.Unmarshal([]byte(r.Header.Get("X-Body")), &params)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	email, err := cache.Get(fmt.Sprintf("email_verify:%s", params.Hash)).Result()
	if err != nil && err != redis.Nil {
		myerrors.ServerError(w, r, err)
		return
	}

	if err == redis.Nil {
		// Invalid email
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	deletedCount, err := cache.Del(fmt.Sprintf("email_verify:%s", params.Hash)).Result()
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	if deletedCount < 1 {
		// Duplicate concurrent requests
		w.WriteHeader(http.StatusConflict)
		return
	}

	userParams := &models.StdParams{Cache: cache, W: w, R: r}
	user := &models.User{
		Email: email,
	}

	err = user.GetByEmail(appengine.NewContext(r))
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

	_, err = user.SetAccountFlag(userParams, models.UserAccountEmailVerified)
	if err != nil {
		myerrors.ServerError(w, r, err)
		return
	}

}
