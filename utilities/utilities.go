package utilities

import (
	"fmt"
	"net/http"
	"os"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/warent/phrhero-calcifer/providers"
)

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
