package order

import (
	"context"
	"mall/adaptor/repo/model"
	"mall/consts"
	"mall/utils/logger"
	"mall/utils/tools"

	"go.uber.org/zap"
)

func IsWaitPay(order *model.Order) bool {
	if order == nil {
		return false
	}
	return order.Status == consts.OrderStatusWaitPay
}

func (s *Service) TimeOutOrderCancel(ctx context.Context, orderID int64) error {
	orderUUID := tools.UUIDHex()
	locked, err := s.rdsOrder.GetOrderLock(ctx, orderID, orderUUID)
	if err != nil {
		logger.Error("TimeOutOrderCancel GetOrderLock error", zap.Error(err), zap.Any("order_id", orderID))
		return err
	}
	if !locked {
		logger.Error("TimeOutOrderCancel other processing", zap.Error(err), zap.Any("order_id", orderID))
		return err
	}
	defer s.rdsOrder.UnLockOrder(ctx, orderID, orderUUID)

	order, err := s.order.GetOrderByID(ctx, orderID)
	if err != nil {
		logger.Error("TimeOutOrderCancel GetOrderByID error", zap.Error(err), zap.Any("order_id", orderID))
		return err
	}
	if IsWaitPay(order) {
		err = s.order.CancelOrder(ctx, orderID, -1, consts.TimeoutCancel)
		if err != nil {
			logger.Error("TimeOutOrderCancel CancelOrder error", zap.Error(err), zap.Any("order_id", orderID))
			return err
		}
	}
	err = s.rdsOrder.DelTimeoutOrderCancel(ctx, orderID)
	if err != nil {
		logger.Error("TimeOutOrderCancel DelTimeoutOrderCancel error", zap.Error(err), zap.Any("order_id", orderID))
	}
	err = s.rdsOrder.DelOrderPayResult(ctx, orderID)
	if err != nil {
		logger.Error("TimeOutOrderCancel DelOrderPayResult error", zap.Error(err), zap.Any("order_id", orderID))
	}
	return nil
}
