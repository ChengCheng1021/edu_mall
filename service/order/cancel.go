package order

import (
	"context"
	"errors"
	"mall/common"
	"mall/consts"
	"mall/utils/logger"
	"mall/utils/tools"

	"github.com/smartwalle/alipay/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (s *Service) cancelOrder(ctx context.Context, orderID, cancelBy int64, cancelType int32) common.Errno {
	orderUUID := tools.UUIDHex()
	locked, err := s.rdsOrder.GetOrderLock(ctx, orderID, orderUUID)
	if err != nil {
		logger.Error("cancelOrder GetOrderLock error", zap.Error(err), zap.Int64("order_id", orderID))
		return common.DatabaseErr.WithErr(err)
	}
	if !locked {
		return common.OrderLockedErr
	}
	defer s.rdsOrder.UnLockOrder(ctx, orderID, orderUUID)

	order, err := s.order.GetOrderByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.OrderNotFoundErr
		}
		logger.Error("cancelOrder GetOrderByID error", zap.Error(err), zap.Int64("order_id", orderID))
		return common.DatabaseErr.WithErr(err)
	}
	if !IsWaitPay(order) {
		return common.OrderCantCancelErr
	}
	if cancelType == consts.CustomerCancel && order.UserID != cancelBy {
		return common.PermissionErr
	}
	if err = s.order.CancelOrder(ctx, orderID, cancelBy, cancelType); err != nil {
		logger.Error("cancelOrder error", zap.Error(err), zap.Int64("order_id", orderID), zap.Int64("cancel_by", cancelBy), zap.Int32("cancel_type", cancelType))
		return common.ServerErr.WithErr(err)
	}
	_ = s.aliPay.CloseOrder(ctx, alipay.TradeClose{
		OutTradeNo: order.InnerTradeNo,
	})
	_ = s.rdsOrder.DelOrderPayResult(ctx, order.ID)
	_ = s.rdsOrder.DelTimeoutOrderCancel(ctx, order.ID)
	return common.OK
}
