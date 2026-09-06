package redis

import (
	"context"
	"errors"
	"mall/adaptor"
	"time"

	"github.com/go-redis/redis"
)

type ILocker interface {
	// TODO 把锁进行uuid处理，保证自己设置的锁，不能被其他地方所解锁
	GetLock(ctx context.Context, lockKey string) (bool, error)
	UnLock(ctx context.Context, lockKey string) error
	AwaitLock(ctx context.Context, lockKey string, timeout time.Duration) error
}

type Locker struct {
	redis *redis.Client
}

func NewLocker(adaptor adaptor.IAdaptor) *Locker {
	return &Locker{
		redis: adaptor.GetRedis(),
	}
}
func (l *Locker) GetLock(ctx context.Context, lockKey string) (bool, error) {
	lockSuccess, err := l.redis.SetNX(lockKey, 1, time.Second*60).Result()
	if err != nil {
		return false, err
	}
	return lockSuccess, nil
}

func (l *Locker) UnLock(ctx context.Context, lockKey string) error {
	_, err := l.redis.Del(lockKey).Result()
	if err != nil {
		return err
	}
	return nil
}

func (l *Locker) AwaitLock(ctx context.Context, lockKey string, timeout time.Duration) error {
	startTime := time.Now()
	for {
		ttl, err := l.redis.TTL(lockKey).Result()
		if err != nil {
			return err
		}
		if int(ttl) >= 0 {
			time.Sleep(time.Millisecond * 100)
		} else {
			return nil
		}
		if time.Now().Sub(startTime) > timeout {
			return errors.New("wait lock timeout")
		}
	}
}
