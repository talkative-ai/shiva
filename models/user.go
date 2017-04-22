package models

import (
	"net/http"

	"fmt"

	"github.com/warent/phrhero-backend/errors"
	"github.com/warent/phrhero-backend/providers"
)

type User struct {
	Email     string
	FirstName string
	LastName  string
}

func (user *User) ValidateNewAccount(w http.ResponseWriter, r *http.Request) {
	client, err := providers.ConnectRedis(r)
	if err != nil {
		errors.ServerError(w, r, err)
		return
	}

	userStatus, err := client.HGet(fmt.Sprintf("user:%s", user.Email), "status").Result()
	if err != nil {
		errors.ServerError(w, r, err)
		return
	}

	if userStatus != "" {
		// User already exists or is registering
	}

	// Register user

}
