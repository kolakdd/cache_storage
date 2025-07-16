package repo

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheRepo interface {
	SetObjectList(key string, value []byte)
	GetObjectList(key string) ([]byte, bool)
	DelObjectList() // удаляет все

	SetDownloadObject(key string, value []byte)
	GetDownloadObject(key string) ([]byte, bool)
	DelDownloadObject(key string) // удаляет выборочно по objID
}

type cacheRepo struct {
	redis    *redis.Client
	cacheDur time.Duration
}

func NewCacheRepo(c *redis.Client) CacheRepo {
	dur := time.Minute * 5
	return &cacheRepo{c, dur}
}

// ObjectList
func (r *cacheRepo) SetObjectList(key string, value []byte) {
	ctx := context.Background()
	key = "cache:objlist:" + key

	r.redis.SetEx(ctx, key, value, r.cacheDur)
}
func (r *cacheRepo) GetObjectList(key string) ([]byte, bool) {
	ctx := context.Background()
	key = "cache:objlist:" + key

	res := r.redis.Get(ctx, key)
	if res.Err() != nil {
		return nil, false
	}
	return []byte(res.Val()), true
}
func (r *cacheRepo) DelObjectList() {
	ctx := context.Background()
	pattern := "cache:objlist:*"
	deletePatern(ctx, r.redis, pattern)
}

// DownloadObject
func (r *cacheRepo) SetDownloadObject(key string, value []byte) {
	ctx := context.Background()
	key = "cache:objdownload:" + key
	r.redis.SetEx(ctx, key, value, r.cacheDur)
}
func (r *cacheRepo) GetDownloadObject(key string) ([]byte, bool) {
	ctx := context.Background()
	key = "cache:objdownload:" + key

	res := r.redis.Get(ctx, key)
	if res.Err() != nil {
		return nil, false
	}
	return []byte(res.Val()), true
}
func (r *cacheRepo) DelDownloadObject(key string) {
	ctx := context.Background()
	pattern := "cache:objdownload:" + key + "*"
	deletePatern(ctx, r.redis, pattern)
}

func deletePatern(ctx context.Context, r *redis.Client, pattern string) {
	var cursor uint64
	for {
		keys, nextCursor, err := r.Scan(ctx, cursor, pattern, 0).Result()
		if err != nil {
			return
		}
		if len(keys) > 0 {
			_, err := r.Del(ctx, keys...).Result()
			if err != nil {
				return
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
}
