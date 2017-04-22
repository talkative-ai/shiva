package models

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/warent/phrhero-backend/errors"
)

// User contains all the properties of the User model. The functions may mutate the model itself and internal storage representations
type User struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

// StdParams Standard parameters used within models for storage mutations and logging
type StdParams struct {
	Cache *redis.Client
	W     http.ResponseWriter
	R     *http.Request
}

// UserAccountStatus is enforces type-safety for what it self-describes
type UserAccountStatus uint8

const (
	// USER_ACCOUNT_DNE User account does not exist
	USER_ACCOUNT_DNE UserAccountStatus = iota

	// USER_ACCOUNT_CREATING User account is in the process of being created
	USER_ACCOUNT_CREATING UserAccountStatus = iota

	// USER_ACCOUNT_EMAIL_PENDING User account has gone through initial registration and is awaiting email conformation
	USER_ACCOUNT_EMAIL_PENDING UserAccountStatus = iota

	// USER_ACCOUNT_OK User account is all set up with no action items
	USER_ACCOUNT_OK UserAccountStatus = iota
)

func (user *User) GetAccountStatus(params *StdParams) (UserAccountStatus, error) {

	accountStatus, err := params.Cache.HGet(fmt.Sprintf("user:%s", user.Email), "account_status").Result()
	if err != nil {
		errors.ServerError(params.W, params.R, err)
		return USER_ACCOUNT_DNE, err
	}

	if accountStatus == "" {
		return USER_ACCOUNT_DNE, nil
	}

	accStatParse, _ := strconv.ParseUint(accountStatus, 10, 8)

	return UserAccountStatus(accStatParse), err
}

func (user *User) SetAccountStatus(params *StdParams, status UserAccountStatus) (bool, error) {

	isNewValue, err := params.Cache.HSet(fmt.Sprintf("user:%s", user.Email), "account_status", status).Result()
	if err != nil {
		errors.ServerError(params.W, params.R, err)
		return false, err
	}

	return isNewValue, err
}
