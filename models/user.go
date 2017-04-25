package models

import (
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"

	"time"

	"net/url"

	"github.com/go-redis/redis"
	"github.com/phrhero/stdapi/phrerrors"
	"github.com/phrhero/stdapi/utilities"
)

// User contains all the properties of the User model. The functions may mutate the model itself and internal storage representations
type User struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Salt      string
}

// StdParams Standard parameters used within for storage mutations and logging
type StdParams struct {
	Cache *redis.Client
	W     http.ResponseWriter
	R     *http.Request
}

// UserAccountFlag is enforces type-safety for what it self-describes
type UserAccountFlag uint32

const (
	// UserAccountExists User account does not exist
	UserAccountExists UserAccountFlag = iota

	// UserAccountFlagEmailVerified Has been verified
	UserAccountEmailVerified UserAccountFlag = iota
)

func (user *User) HasAccountFlag(params *StdParams, flag UserAccountFlag) (bool, error) {

	isMember, err := params.Cache.SIsMember(fmt.Sprintf("user:%s:flags", user.Email), uint32(flag)).Result()
	if err != nil && err.Error() != "redis: nil" {
		phrerrors.ServerError(params.W, params.R, err)
		return false, err
	}

	return isMember, nil

}

func (user *User) SetAccountFlag(params *StdParams, flag UserAccountFlag) (int64, error) {

	newValueCount, err := params.Cache.SAdd(fmt.Sprintf("user:%s:flags", user.Email), uint32(flag)).Result()
	if err != nil {
		phrerrors.ServerError(params.W, params.R, err)
		return 0, err
	}

	return newValueCount, err
}

func (user *User) RegisterAccount(params *StdParams) error {
	ctx := appengine.NewContext(params.R)

	k := datastore.NewKey(ctx, "User", "0", 0, nil)

	if _, err := datastore.Put(ctx, k, user); err != nil {
		return err
	}

	return nil
}

type SendVerificationEmailStatus int8

const (
	EMAIL_COOLDOWN    SendVerificationEmailStatus = iota
	ALREADY_VERIFIED  SendVerificationEmailStatus = iota
	VERIFICATION_SENT SendVerificationEmailStatus = iota
)

func (user *User) SendVerificationEmail(params *StdParams) (SendVerificationEmailStatus, error) {

	isVerified, err := user.HasAccountFlag(params, UserAccountEmailVerified)
	if err != nil {
		return -1, err
	}

	// User account has already been verified
	if isVerified {
		return ALREADY_VERIFIED, nil
	}

	cooldownRefreshed, err := params.Cache.SetNX(fmt.Sprintf("user:%s:cooldown:email_sent", user.Email), 1, time.Second*30).Result()
	if err != nil {
		return -1, err
	}

	if !cooldownRefreshed {
		return EMAIL_COOLDOWN, nil
	}

	ctx := appengine.NewContext(params.R)

	t := taskqueue.NewPOSTTask("/v1/email/verification", url.Values{
		"Email": []string{user.Email},
	})

	if _, err := taskqueue.Add(ctx, t, "kiki"); err != nil {
		log.Errorf(ctx, err.Error())
	}

	return VERIFICATION_SENT, nil
}

func (user *User) EncryptPassword() error {

	salt, err := utilities.GenerateRandomString(32)
	if err != nil {
		return err
	}

	password := fmt.Sprintf("%s%s%s", salt, user.Password, salt)

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hash)
	user.Salt = salt

	return nil
}
