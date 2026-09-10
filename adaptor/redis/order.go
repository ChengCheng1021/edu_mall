package redis

import (
	"context"
	"fmt"
	"mall/adaptor"
	"mall/config"
	"time"

	"github.com/go-redis/redis"
)

type IOrder interface {
	SetOrderCalcFee(ctx context.Context, feeUUID string, feeData string, expire time.Duration) error
	GetOrderCalcFee(ctx context.Context, feeUUID string) (string, error)
}

type Order struct {
	redisClient *redis.Client
}

func NewOrder(adaptor adaptor.IAdaptor) *Order {
	return &Order{
		redisClient: adaptor.GetRedis(),
	}
}

func fmtOrderCalcFeeKey(feeUUID string) string {
	return fmt.Sprintf("%s:order:calc:fee:%s", config.ServerFullName, feeUUID)
}

func fmtOrderLockKey(orderId int64) string {
	return fmt.Sprintf("%s:lock:order:%d", config.ServerFullName, orderId)
}

func (o *Order) SetOrderCalcFee(ctx context.Context, feeUUID string, feeData string, expire time.Duration) error {
	return o.redisClient.Set(fmtOrderCalcFeeKey(feeUUID), feeData, expire).Err()
}
func (o *Order) GetOrderCalcFee(ctx context.Context, feeUUID string) (string, error) {
	return o.redisClient.Get(fmtOrderCalcFeeKey(feeUUID)).Result()
}
