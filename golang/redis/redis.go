package redis

import (
	"github.com/kolakdd/cache_storage/golang/repo"
	"github.com/redis/go-redis/v9"
)

func InitRedis(envRepo repo.RepositoryEnv) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: envRepo.GetRegisURL(),
		DB:   0,
	})
	return client
}
