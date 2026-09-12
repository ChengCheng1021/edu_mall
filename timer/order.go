package timer

import (
	"context"
	"mall/utils/logger"
	"time"

	"go.uber.org/zap"
)

func (s *Service) OrderTimeOutCancel() {
	ctx := context.TODO()
	defer func() {
		if err := recover(); err != nil {
			logger.Error("OrderTimeOutCancel panic", zap.Any("err", err))
		}
	}()
	orderList, err := s.rdsOrder.GetTimeWaitOrderCancel(ctx)
	if err != nil {
		logger.Error("OrderTimeOutCancel GetTimeWaitOrderCancel error", zap.Error(err))
		return
	}
	for _, orderID := range orderList {
		err = s.order.TimeOutOrderCancel(ctx, orderID)
		if err != nil {
			logger.Error("OrderTimeOutCancel TimeOutOrderCancel error", zap.Error(err), zap.Int64("order_id", orderID))
			continue
		}
	}
}

func (s *Service) OrderPayResultQuery() {
	ctx := context.TODO()
	defer func() {
		if err := recover(); err != nil {
			logger.Error("OrderPayResultQuery panic", zap.Any("err", err))
		}
	}()
	orderMap, err := s.rdsOrder.GetOrderPayResult(ctx)
	if err != nil {
		logger.Error("OrderPayResultQuery GetOrderPayResult error", zap.Error(err))
		return
	}
	for orderID, orderTime := range orderMap {
		err = s.order.QueryOrderPayResult(ctx, orderID, orderTime)
		if err != nil {
			logger.Error("OrderPayResultQuery GetOrderPayResult error",
				zap.Error(err), zap.Int64("order_id", orderID), zap.Time("order_time", time.UnixMilli(orderTime)))
			continue
		}
	}
}

func (s *Service) OrderRefundQuery() {
	ctx := context.TODO()
	defer func() {
		if err := recover(); err != nil {
			logger.Error("OrderRefundQuery panic", zap.Any("err", err))
		}
	}()
	orderMap, err := s.rdsOrder.GetOrderRefundResult(ctx)
	if err != nil {
		logger.Error("OrderRefundQuery GetOrderRefundResult error", zap.Error(err))
		return
	}
	for orderID, orderTime := range orderMap {
		err = s.order.QueryOrderRefundResult(ctx, orderID, orderTime)
		if err != nil {
			logger.Error("OrderRefundQuery QueryOrderRefundResult error",
				zap.Error(err), zap.Int64("order_id", orderID), zap.Time("order_time", time.UnixMilli(orderTime)))
			continue
		}
	}
}
