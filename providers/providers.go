package providers

import (
	"net"
	"os"

	"net/http"

	"github.com/go-redis/redis"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/socket"
)

// ConnectRedis dials and returns a Redis client to query against, or abort with error
func ConnectRedis(r *http.Request) (*redis.Client, error) {

	ctx := appengine.NewContext(r)

	client := redis.NewClient(&redis.Options{
		Dialer: func() (net.Conn, error) {
			conn, err := socket.Dial(ctx, "tcp", os.Getenv("REDIS_ADDR"))
			if err != nil {
				log.Criticalf(ctx, err.Error())
				return nil, err
			}
			return conn, nil
		},
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	_, err := client.Ping().Result()
	if err != nil {
		log.Criticalf(ctx, err.Error())
		return nil, err
	}

	return client, nil

}
