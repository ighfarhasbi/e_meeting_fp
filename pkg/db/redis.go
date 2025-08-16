package db

import "github.com/redis/go-redis/v9"

func NewRedis(redisUrl string) (*redis.Client, error) {
	return redis.NewClient(&redis.Options{
		Addr:     redisUrl,
		Password: "",
		DB:       0,
	}), nil
}
