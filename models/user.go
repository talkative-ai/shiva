package models

import (
	"fmt"
	"net/http"
	"strconv"

	"time"

	"github.com/go-redis/redis"
	"github.com/phrhero/stdapi/phrerrors"
)

// User contains all the properties of the User model. The functions may mutate the model itself and internal storage representations
type User struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

// StdParams Standard parameters used within for storage mutations and logging
type StdParams struct {
	Cache *redis.Client
	W     http.ResponseWriter
	R     *http.Request
}

// UserAccountStatus is enforces type-safety for what it self-describes
type UserAccountStatus uint32

const (
	// USER_ACCOUNT_DNE User account does not exist
	USER_ACCOUNT_DNE UserAccountStatus = 1 << iota

	// USER_ACCOUNT_CREATING User account is in the process of being created
	USER_ACCOUNT_CREATING UserAccountStatus = 1 << iota

	// USER_ACCOUNT Has been verified
	USER_ACCOUNT_EMAIL_VERIFIED UserAccountStatus = 1 << iota

	// USER_ACCOUNT_OK User account is all set up with no action items
	USER_ACCOUNT_OK UserAccountStatus = 1 << iota
)

func (user *User) GetAccountStatus(params *StdParams) (UserAccountStatus, error) {

	accountStatus, err := params.Cache.HGet(fmt.Sprintf("user:%s", user.Email), "account_status").Result()
	if err != nil && err.Error() != "redis: nil" {
		phrerrors.ServerError(params.W, params.R, err)
		return USER_ACCOUNT_DNE, err
	}

	if accountStatus == "" {
		return USER_ACCOUNT_DNE, nil
	}

	accStatParse, _ := strconv.ParseUint(accountStatus, 10, 8)

	return UserAccountStatus(accStatParse), err
}

func (user *User) SetAccountStatus(params *StdParams, status UserAccountStatus) (bool, error) {

	isNewValue, err := params.Cache.HSet(fmt.Sprintf("user:%s", user.Email), "account_status", uint32(status)).Result()
	if err != nil {
		phrerrors.ServerError(params.W, params.R, err)
		return false, err
	}

	return isNewValue, err
}

func (user *User) Save(params *StdParams) (bool, error) {
	return false, nil
}

type SendVerificationEmailStatus int8

const (
	EMAIL_COOLDOWN    SendVerificationEmailStatus = iota
	ALREADY_VERIFIED  SendVerificationEmailStatus = iota
	VERIFICATION_SENT SendVerificationEmailStatus = iota
)

func (user *User) SendVerificationEmail(params *StdParams) (SendVerificationEmailStatus, error) {

	accStatus, err := user.GetAccountStatus(params)
	if err != nil {
		return -1, err
	}

	// User account has already been verified
	if accStatus&USER_ACCOUNT_EMAIL_VERIFIED != 0 {
		return ALREADY_VERIFIED, nil
	}

	cooldownRefreshed, err := params.Cache.SetNX(fmt.Sprintf("user:%s:cooldown:email_sent", user.Email), 1, time.Second*30).Result()
	if err != nil {
		return -1, err
	}

	if !cooldownRefreshed {
		return EMAIL_COOLDOWN, nil
	}

	return VERIFICATION_SENT, nil
}
