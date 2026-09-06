package redis

import (
	"context"
	"fmt"
	"mall/adaptor"
	"mall/config"
	"time"

	"github.com/go-redis/redis"
	"github.com/gogf/gf/util/gconv"
)

type IAccessToken interface {
	SetAccessToken(ctx context.Context, cacheKey string, token string, expire time.Duration) error
	GetAccessToken(ctx context.Context, cacheKey string) (string, int64, error)
}

func cacheTokenKeyFmt(appCode int32) string {
	return fmt.Sprintf("%s:cachetoken:%d", config.ServerFullName, appCode)
}

func lockTokenKeyFmt(appCode int32) string {
	return fmt.Sprintf("%s:locktoken:%d", config.ServerFullName, appCode)
}

type AccessToken struct {
	redis *redis.Client
}

func NewAccessToken(adaptor adaptor.IAdaptor) *AccessToken {
	return &AccessToken{
		redis: adaptor.GetRedis(),
	}
}

func (a *AccessToken) SetAccessToken(ctx context.Context, cacheKey string, token string, expire time.Duration) error {
	return a.redis.Set(cacheKey, token, expire).Err()
}
func (a *AccessToken) GetAccessToken(ctx context.Context, cacheKey string) (string, int64, error) {
	tokenJson, err := a.redis.Get(cacheKey).Result()
	if err != nil {
		return "", 0, err
	}
	expireIn, err := a.redis.TTL(cacheKey).Result()
	if err != nil {
		return "", 0, err
	}
	return tokenJson, gconv.Int64(expireIn.Seconds()), nil
}
