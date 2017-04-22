package utilities

import (
	"fmt"
	"net/http"
	"os"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"

	mailgun "gopkg.in/mailgun/mailgun-go.v1"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis"
	"github.com/warent/phrhero-calcifer/providers"
)

// StdParams Standard parameters used within for storage mutations and logging
type StdParams struct {
	Cache *redis.Client
	W     http.ResponseWriter
	R     *http.Request
}

// ParseJTWClaims parses a JWT token for value accessing
func ParseJTWClaims(tokenString string) (map[string]interface{}, error) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_KEY")), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, err
}

// ValidateJWT ensures the validity of a JWT
// Used by JWT prehandler and by post.user.validate
func ValidateJWT(r *http.Request, token string) bool {

	if token == "" {
		return false
	}

	cache, err := providers.ConnectRedis(r)

	if err != nil {
		return false
	}

	_, err = cache.Get("jwt:" + token).Result()

	cache.Close()

	return err == nil
}

func SendVerificationEmail(params *StdParams, email string) {
	mg := mailgun.NewMailgun(os.Getenv("MG_DOMAIN"), os.Getenv("MG_API_KEY"), os.Getenv("MG_PUBLIC_API_KEY"))
	ctx := appengine.NewContext(params.R)
	client := urlfetch.Client(ctx)
	mg.SetClient(client)

	message := mailgun.NewMessage(
		"no-reply@phrhero.com",
		"phrhero Email Verification!",
		"Please verify your email here",
		email)
	_, _, err := mg.Send(message)
	if err != nil {
		log.Errorf(ctx, err.Error())
	}
}
