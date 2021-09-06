package redis

import (
	"github.com/go-redis/redis"
	"time"
)

type RedisService struct {
	client *redis.Client
}

type IRedisService interface {
	SetObj(key string, obj string) error
	GetObj(key string) string
}

func NewRedisService(addr string) *RedisService {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return &RedisService{
		client: client,
	}
}

func (svc *RedisService) SetObj(key string, obj string) error {
	err := svc.client.Set(key, obj, time.Hour).Err()
	return err
}

func (svc *RedisService) GetObj(key string) string {
	obj := svc.client.Get(key)
	if obj != nil {
		return obj.Val()
	}
	return ""
}
