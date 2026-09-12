package redis

import (
	"context"
	"errors"
	"fmt"
	"mall/adaptor"
	"mall/config"
	"time"

	"github.com/go-redis/redis"
	"github.com/gogf/gf/util/gconv"
)

type IOrder interface {
	SetOrderCalcFee(ctx context.Context, feeUUID string, feeData string, expire time.Duration) error
	GetOrderCalcFee(ctx context.Context, feeUUID string) (string, error)

	// 订单分布式锁
	GetOrderLock(ctx context.Context, orderID int64, uuid string) (bool, error)
	UnLockOrder(ctx context.Context, orderID int64, uuid string) error

	// 订单超时取消
	SetTimeoutOrderCancel(ctx context.Context, orderID int64) error
	GetTimeWaitOrderCancel(ctx context.Context) ([]int64, error)
	DelTimeoutOrderCancel(ctx context.Context, orderID int64) error

	// 订单支付结果轮询
	SetOrderPayResult(ctx context.Context, orderID int64) error
	GetOrderPayResult(ctx context.Context) (map[int64]int64, error)
	DelOrderPayResult(ctx context.Context, orderID int64) error
	CheckInPayment(ctx context.Context, orderID int64) (bool, error)

	// 订单退款结果查询
	SetOrderRefundResult(ctx context.Context, orderID int64) error
	GetOrderRefundResult(ctx context.Context) (map[int64]int64, error)
	DelOrderRefundResult(ctx context.Context, orderID int64) error
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

func fmtOrderLockKey(orderID int64) string {
	return fmt.Sprintf("%s:lock:order:%d", config.ServerFullName, orderID)
}
func (o *Order) GetOrderLock(ctx context.Context, orderID int64, uuid string) (bool, error) {
	redisKey := fmtOrderLockKey(orderID)
	return o.redisClient.SetNX(redisKey, uuid, time.Second*10).Result()
}
func (o *Order) UnLockOrder(ctx context.Context, orderID int64, uuid string) error {
	redisKey := fmtOrderLockKey(orderID)
	return luaUnlock.Run(o.redisClient, []string{redisKey}, uuid).Err()
}

func (o *Order) SetOrderCalcFee(ctx context.Context, feeUUID string, feeData string, expire time.Duration) error {
	return o.redisClient.Set(fmtOrderCalcFeeKey(feeUUID), feeData, expire).Err()
}
func (o *Order) GetOrderCalcFee(ctx context.Context, feeUUID string) (string, error) {
	return o.redisClient.Get(fmtOrderCalcFeeKey(feeUUID)).Result()
}

// 订单超时取消
const (
	TimeoutOrderCancelExpire = -1 * time.Minute * 30
)

func fmtTimeoutOrderCancelZSetKey() string {
	return fmt.Sprintf("%s:order:timeout:cancel", config.ServerFullName)
}
func (o *Order) SetTimeoutOrderCancel(ctx context.Context, orderID int64) error {
	redisKey := fmtTimeoutOrderCancelZSetKey()
	_, err := o.redisClient.ZAdd(redisKey, redis.Z{
		Score:  float64(time.Now().UnixMilli()),
		Member: orderID,
	}).Result()
	if err != nil {
		return err
	}
	return nil
}

func (o *Order) GetTimeWaitOrderCancel(ctx context.Context) ([]int64, error) {
	redisKey := fmtTimeoutOrderCancelZSetKey()
	timeNow := time.Now()
	minTime := timeNow.AddDate(0, 0, -7).UnixMilli()
	maxTime := timeNow.Add(TimeoutOrderCancelExpire).UnixMilli()
	strList, err := o.redisClient.ZRangeByScore(redisKey, redis.ZRangeBy{
		Min:    gconv.String(minTime),
		Max:    gconv.String(maxTime),
		Offset: 0,
		Count:  500,
	}).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return []int64{}, nil
		}
		return nil, err
	}
	retOrderList := make([]int64, 0)
	for _, str := range strList {
		orderID := gconv.Int64(str)
		if orderID > 0 {
			retOrderList = append(retOrderList, orderID)
		}
	}
	return retOrderList, nil
}
func (o *Order) DelTimeoutOrderCancel(ctx context.Context, orderID int64) error {
	redisKey := fmtTimeoutOrderCancelZSetKey()
	_, err := o.redisClient.ZRem(redisKey, orderID).Result()
	return err
}

func fmtOrderPayResultHashKey() string {
	return fmt.Sprintf("%s:order:pay:result", config.ServerFullName)
}

// 订单支付结果轮询
func (o *Order) SetOrderPayResult(ctx context.Context, orderID int64) error {
	redisKey := fmtOrderPayResultHashKey()
	_, err := o.redisClient.HSet(redisKey, gconv.String(orderID), time.Now().UnixMilli()).Result()
	return err
}
func (o *Order) GetOrderPayResult(ctx context.Context) (map[int64]int64, error) {
	redisKey := fmtOrderPayResultHashKey()
	getMap, err := o.redisClient.HGetAll(redisKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return map[int64]int64{}, nil
		}
		return nil, err
	}
	retMap := make(map[int64]int64)
	for k, v := range getMap {
		retMap[gconv.Int64(k)] = gconv.Int64(v)
	}
	return retMap, nil
}
func (o *Order) DelOrderPayResult(ctx context.Context, orderID int64) error {
	redisKey := fmtOrderPayResultHashKey()
	_, err := o.redisClient.HDel(redisKey, gconv.String(orderID)).Result()
	return err
}

func (o *Order) CheckInPayment(ctx context.Context, orderID int64) (bool, error) {
	return o.redisClient.HExists(fmtOrderPayResultHashKey(), gconv.String(orderID)).Result()
}

func fmtOrderRefundResultHashKey() string {
	return fmt.Sprintf("%s:order:refund:result", config.ServerFullName)
}

func (o *Order) SetOrderRefundResult(ctx context.Context, orderID int64) error {
	redisKey := fmtOrderRefundResultHashKey()
	_, err := o.redisClient.HSet(redisKey, gconv.String(orderID), time.Now().UnixMilli()).Result()
	return err
}
func (o *Order) GetOrderRefundResult(ctx context.Context) (map[int64]int64, error) {
	redisKey := fmtOrderRefundResultHashKey()
	getMap, err := o.redisClient.HGetAll(redisKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return map[int64]int64{}, nil
		}
		return nil, err
	}
	retMap := make(map[int64]int64)
	for k, v := range getMap {
		retMap[gconv.Int64(k)] = gconv.Int64(v)
	}
	return retMap, nil
}
func (o *Order) DelOrderRefundResult(ctx context.Context, orderID int64) error {
	redisKey := fmtOrderRefundResultHashKey()
	_, err := o.redisClient.HDel(redisKey, gconv.String(orderID)).Result()
	return err
}
