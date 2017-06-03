package providers

import (
	"fmt"
	"os"

	"github.com/go-redis/redis"
)

// ConnectRedis dials and returns a Redis client to query against, or abort with error
func ConnectRedis() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	_, err := client.Ping().Result()
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return client, nil

}
