package order

import (
	"context"
	"encoding/json"
	"errors"
	"mall/adaptor/repo/model"
	"mall/consts"
	"mall/service/do"
	"mall/utils/logger"
	"mall/utils/tools"
	"time"

	"github.com/gogf/gf/util/gconv"
	"github.com/smartwalle/alipay/v3"
	"go.uber.org/zap"
)

const (
	PaySuccess = "TRADE_SUCCESS"
)

func (s *Service) QueryOrderPayResult(ctx context.Context, orderID, orderTime int64) error {
	if time.Now().Sub(time.UnixMilli(orderTime)).Minutes() > 5 {
		err := s.rdsOrder.DelOrderPayResult(ctx, orderID)
		if err != nil {
			logger.Error("TimeOutOrderCancel DelOrderPayResult error", zap.Error(err), zap.Any("order_id", orderID))
			return err
		}
		return nil
	}
	orderUUID := tools.UUIDHex()
	locked, err := s.rdsOrder.GetOrderLock(ctx, orderID, orderUUID)
	if err != nil {
		logger.Error("QueryOrderPayResult GetOrderLock error", zap.Error(err), zap.Any("order_id", orderID))
		return err
	}
	if !locked {
		logger.Error("QueryOrderPayResult other processing", zap.Error(err), zap.Any("order_id", orderID))
		return err
	}
	defer s.rdsOrder.UnLockOrder(ctx, orderID, orderUUID)

	order, err := s.order.GetOrderByID(ctx, orderID)
	if err != nil {
		logger.Error("QueryOrderPayResult GetOrderByID error", zap.Error(err), zap.Any("order_id", orderID))
		return err
	}
	if IsWaitPay(order) {
		err = s.handlerOrderPayResult(ctx, order, orderTime)
		if err != nil {
			logger.Error("QueryOrderPayResult handlerOrderPayResult error", zap.Error(err), zap.Any("order_id", orderID))
			return err
		}
	}
	err = s.rdsOrder.DelTimeoutOrderCancel(ctx, orderID)
	if err != nil {
		logger.Error("QueryOrderPayResult DelTimeoutOrderCancel error", zap.Error(err), zap.Any("order_id", orderID))
	}
	err = s.rdsOrder.DelOrderPayResult(ctx, orderID)
	if err != nil {
		logger.Error("TimeOutOrderCancel DelOrderPayResult error", zap.Error(err), zap.Any("order_id", orderID))
	}
	return nil
}

func IsPayed(state string) bool {
	return state == PaySuccess
}
func (s *Service) handlerOrderPayResult(ctx context.Context, order *model.Order, orderTime int64) error {
	resp, err := s.aliPay.QueryOrderByOutTradeNo(ctx, alipay.TradeQuery{
		OutTradeNo: order.InnerTradeNo,
	})
	if err != nil {
		logger.Error("handlerOrderPayResult QueryOrderByOutTradeNo error", zap.Error(err), zap.Any("order_id", order.ID))
		return err
	}
	if !IsPayed(gconv.String(resp.TradeStatus)) {
		return errors.New("order not payment")
	}
	paymentTime, err := time.Parse(time.RFC3339, resp.SendPayDate)
	if err != nil {
		paymentTime = time.Now()
	}
	benefitFunc := func() error {
		return s.userBenefitPackage(ctx, order, paymentTime)
	}
	err = s.order.UpdateOrderPaySuccess(ctx, &do.UpdateOrderPaySuccess{
		OrderID:       order.ID,
		PaymentAt:     paymentTime,
		TransactionID: resp.TradeNo,
		BenefitFunc:   benefitFunc,
	})
	if err != nil {
		logger.Error("TimeOutOrderCancel UpdateOrderPaySuccess error", zap.Error(err), zap.Any("order_id", order.ID))
		return err
	}
	return nil
}
func (s *Service) userBenefitPackage(ctx context.Context, order *model.Order, paymentTime time.Time) error {
	orderItems, err := s.order.GetOrderItems(ctx, order.ID)
	if err != nil {
		logger.Error("userBenefitPackage GetOrderItems error", zap.Error(err), zap.Any("order_id", order.ID))
		return err
	}
	buyGoods := make([]do.BuyCourseGoods, 0)
	for _, item := range orderItems {
		courseGoods := &model.CourseGood{}
		err = json.Unmarshal([]byte(item.GoodsSnap), courseGoods)
		if err != nil {
			logger.Error("userBenefitPackage Unmarshal error", zap.Error(err), zap.Any("order_id", order.ID))
			return err
		}
		learnTime := consts.GetExpireTime(courseGoods.LearnTime)
		serviceTime := consts.GetExpireTime(courseGoods.ServiceTime)
		if learnTime == 0 || serviceTime == 0 {
			return errors.New("invalid expire time")
		}
		buyGoods = append(buyGoods, do.BuyCourseGoods{
			OrderItemID:       item.ID,
			GoodsID:           item.GoodsID,
			GoodType:          item.GoodsType,
			LearnExpireTime:   learnTime,
			ServiceExpireTime: serviceTime,
		})
	}
	err = s.userCourse.CreateUserCourse(ctx, &do.CreateUserCourse{
		UserId:     order.UserID,
		OrderID:    order.ID,
		BuyTime:    paymentTime.UnixMilli(),
		CourseList: buyGoods,
	})
	if err != nil {
		logger.Error("userBenefitPackage CreateUserCourse error", zap.Error(err), zap.Any("order_id", order.ID))
		return err
	}
	return nil
}
