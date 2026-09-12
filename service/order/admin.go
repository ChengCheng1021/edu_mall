package order

import (
	"context"
	"encoding/json"
	"errors"
	"mall/adaptor/repo/model"
	"mall/common"
	"mall/consts"
	"mall/service/do"
	"mall/service/dto"
	"mall/utils/logger"
	"mall/utils/tools"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/util/gconv"
	"github.com/samber/lo"
	"github.com/smartwalle/alipay/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (s *Service) GetOrderList(ctx context.Context, req *dto.GetOrderListReq) (*dto.GetOrderListResp, common.Errno) {
	statusList := make([]int32, 0)
	if req.StatusList != "" {
		for _, item := range strings.Split(req.StatusList, ",") {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}
			statusValue, err := strconv.ParseInt(item, 10, 32)
			if err != nil {
				return nil, common.ParamErr.WithMsg("status_list 参数错误")
			}
			statusList = append(statusList, int32(statusValue))
		}
	}
	if len(statusList) == 0 && req.Status != 0 {
		statusList = append(statusList, req.Status)
	}
	list, count, err := s.order.GetOrderList(ctx, &do.GetOrderList{
		Pager:        req.Pager,
		Status:       req.Status,
		StatusList:   statusList,
		OrderID:      req.OrderID,
		UserID:       req.UserID,
		GoodsNameKw:  req.GoodsNameKw,
		CreateStart:  req.CreateStart,
		CreateEnd:    req.CreateEnd,
		PaymentStart: req.PaymentStart,
		PaymentEnd:   req.PaymentEnd,
		RefundStart:  req.RefundStart,
		RefundEnd:    req.RefundEnd,
	})
	if err != nil {
		logger.Error("GetOrderList  GetOrderList error", zap.Error(err), zap.Any("req", req))
		return nil, common.DatabaseErr.WithErr(err)
	}
	return &dto.GetOrderListResp{
		List:  s.convertModelOrderToOrderDto(ctx, list),
		Total: count,
		Pager: req.Pager,
	}, common.OK
}

func (s *Service) enableRefund(order *model.Order) bool {
	enableList := []int32{
		consts.OrderStatusPayed,
		consts.OrderStatusShipped,
		consts.OrderStatusReceived,
		consts.OrderStatusCompleted,
	}
	return lo.Contains(enableList, order.Status)
}

func (s *Service) OrderRefund(ctx context.Context, user *common.AdminUser, req *dto.OrderRefundReq) common.Errno {
	orderUUID := tools.UUIDHex()
	locked, err := s.rdsOrder.GetOrderLock(ctx, req.OrderID, orderUUID)
	if err != nil {
		logger.Error("OrderRefund GetOrderLock error", zap.Error(err), zap.Any("req", req))
		return common.ServerErr.WithErr(err)
	}
	if !locked {
		logger.Error("OrderRefund other processing", zap.Error(err), zap.Any("req", req))
		return common.OrderLockedErr
	}
	defer s.rdsOrder.UnLockOrder(ctx, req.OrderID, orderUUID)

	order, err := s.order.GetOrderByID(ctx, req.OrderID)
	if err != nil {
		logger.Error("OrderRefund GetOrderByID error", zap.Error(err), zap.Any("req", req))
		return common.DatabaseErr.WithErr(err)
	}
	if !s.enableRefund(order) {
		return common.OrderCantRefundErr
	}

	orderItems, err := s.order.GetOrderItems(ctx, req.OrderID)
	if err != nil {
		logger.Error("OrderRefund GetOrderItems error", zap.Error(err), zap.Any("req", req))
		return common.DatabaseErr.WithErr(err)
	}
	orderItemMap := lo.SliceToMap(orderItems, func(item *model.OrderItem) (int64, *model.OrderItem) {
		return item.ID, item
	})
	selectedItemIDs := lo.Uniq(req.ItemIds)
	if len(selectedItemIDs) == 0 {
		selectedItemIDs = lo.Map(orderItems, func(item *model.OrderItem, index int) int64 {
			return item.ID
		})
	}
	for _, itemID := range selectedItemIDs {
		if _, ok := orderItemMap[itemID]; !ok {
			return common.ParamErr.WithMsg("退款商品不存在")
		}
	}
	selectedRefundAmount := lo.SumBy(selectedItemIDs, func(itemID int64) int64 {
		if orderItemMap[itemID] == nil {
			return 0
		}
		return orderItemMap[itemID].PaymentAmount
	})
	if req.Amount <= 0 || req.Amount > selectedRefundAmount {
		return common.OrderRefundAmountErr.WithMsg("退款金额不能超过所选商品实付金额")
	}

	refundReq, outTradeNo, errno := s.assemblyAliRefundParam(ctx, order, req)
	if errno.NotOk() {
		logger.Error("OrderRefund assemblyAliRefundParam error", zap.Error(err), zap.Any("req", req))
		return errno
	}
	handleFun := func(ctx context.Context) error {
		_, err := s.aliPay.RefundOrder(ctx, refundReq)
		if err != nil {
			logger.Error("OrderRefund RefundOrder error", zap.Error(err), zap.Any("req", req))
			return err
		}
		return nil
	}
	// 添加到轮询查退款是否成功的队列里面
	err = s.rdsOrder.SetOrderRefundResult(ctx, req.OrderID)
	if err != nil {
		logger.Error("OrderRefund SetOrderRefundResult error", zap.Error(err), zap.Any("req", req))
		return common.RedisErr.WithErr(err)
	}
	_, err = s.order.CreatOrderRefund(ctx, &do.OrderRefund{
		UserID:      order.UserID,
		OrderID:     req.OrderID,
		ItemIds:     selectedItemIDs,
		Reason:      req.Reason,
		Amount:      req.Amount,
		AdminUserID: user.UserID,
		OutTradeNo:  outTradeNo,
		RefundFun:   handleFun,
	})
	if err != nil {
		// 如果失败了，需要删除刚刚加入的轮询队列
		s.rdsOrder.DelOrderRefundResult(ctx, req.OrderID)
		if strings.Contains(err.Error(), "NOT_ENOUGH") {
			return common.OrderRefundNotEnoughErr
		}
		logger.Error("OrderRefund OrderRefund error", zap.Error(err), zap.Any("req", req))
		return common.DatabaseErr.WithErr(err)
	}
	return common.OK
}

func (s *Service) assemblyAliRefundParam(ctx context.Context, order *model.Order, req *dto.OrderRefundReq) (alipay.TradeRefund, string, common.Errno) {
	_ = ctx
	refundRsp := alipay.TradeRefund{}
	amount := req.Amount
	if amount <= 0 || amount > order.PaymentAmount {
		return refundRsp, "", common.OrderRefundAmountErr
	}

	if req.Reason == "" {
		req.Reason = "用户要求退款"
	}
	suffixAmount := strings.TrimSuffix(gconv.String(amount), "00")
	outTradeNo := tools.UUIDHex()
	refundRsp = alipay.TradeRefund{
		OutTradeNo:   order.InnerTradeNo,
		RefundAmount: suffixAmount,
		RefundReason: req.Reason,
	}
	return refundRsp, outTradeNo, common.OK
}

func (s *Service) QueryOrderRefundResult(ctx context.Context, orderID, orderTime int64) error {
	orderUUID := tools.UUIDHex()
	locked, err := s.rdsOrder.GetOrderLock(ctx, orderID, orderUUID)
	if err != nil {
		logger.Error("QueryOrderRefundResult GetOrderLock error", zap.Error(err), zap.Any("order_id", orderID))
		return err
	}
	if !locked {
		logger.Error("QueryOrderRefundResult other processing", zap.Error(err), zap.Any("order_id", orderID))
		return err
	}
	defer s.rdsOrder.UnLockOrder(ctx, orderID, orderUUID)

	orderRefund, err := s.order.GetOrderRefundByOrderID(ctx, orderID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("QueryOrderRefundResult GetOrderRefundByOrderID error", zap.Error(err), zap.Any("order_id", orderID))
		return err
	}
	if orderRefund == nil {
		err = s.rdsOrder.DelOrderRefundResult(ctx, orderID)
		if err != nil {
			logger.Error("QueryOrderRefundResult DelOrderRefundResult error", zap.Error(err), zap.Any("order_id", orderID))
			return err
		}
		return nil
	}
	if orderRefund.Status != consts.RefundStatusProcessing {
		return s.rdsOrder.DelOrderRefundResult(ctx, orderID)
	}
	order, err := s.order.GetOrderByID(ctx, orderRefund.OrderID)
	if err != nil {
		logger.Error("QueryOrderRefundResult GetOrderByID error", zap.Error(err), zap.Any("order_id", orderRefund.OrderID))
		return err
	}
	aliResp, err := s.aliPay.QueryRefund(ctx, alipay.TradeFastPayRefundQuery{
		OutTradeNo:   order.InnerTradeNo,
		OutRequestNo: order.InnerTradeNo,
	})
	if err != nil {
		logger.Error("QueryOrderRefundResult QueryOrderRefund error", zap.Error(err), zap.Any("order_id", orderID))
		return err
	}
	// 常量只更改了成功,其余未变
	refundStatus := consts.GetRefundStatus(aliResp.RefundStatus)
	if refundStatus == consts.RefundStatusProcessing {
		return nil
	}
	successTime, err := time.Parse(time.RFC3339, aliResp.GMTRefundPay)
	if err != nil {
		successTime = time.Now()
	}
	itemIds := make([]int64, 0)
	err = json.Unmarshal([]byte(orderRefund.ItemIds), &itemIds)
	if err != nil {
		logger.Error("QueryOrderRefundResult Unmarshal error", zap.Error(err), zap.Any("order_id", orderID))
		return err
	}
	handleFun := func(ctx context.Context) error {
		return s.userCourse.DeleteUserCourse(ctx, &do.DeleteUserCourse{
			UserID:       orderRefund.UserID,
			OrderID:      orderRefund.OrderID,
			OrderItemIds: itemIds,
		})
	}
	err = s.order.UpdateOrderRefundResult(ctx, &do.OrderRefundResult{
		OrderRefundID:      orderRefund.ID,
		OrderID:            orderRefund.OrderID,
		RefundID:           aliResp.TradeNo,
		Status:             refundStatus,
		OrderPaymentAmount: order.PaymentAmount,
		SuccessTime:        successTime.UnixMilli(),
		RefundDeliveryFun:  handleFun,
	})
	if err != nil {
		logger.Error("QueryOrderRefundResult OrderRefundResult error", zap.Error(err), zap.Any("order_id", orderID))
		return common.DatabaseErr.WithErr(err)
	}
	return s.rdsOrder.DelOrderRefundResult(ctx, orderID)
}

func (s *Service) AdminCancelOrder(ctx context.Context, user *common.AdminUser, req *dto.CancelOrderReq) common.Errno {
	if user == nil {
		return common.AuthErr
	}
	return s.cancelOrder(ctx, req.OrderID, user.UserID, consts.AdminCancel)
}

func (s *Service) OrderStatistic(ctx context.Context, req *dto.OrderStatReq) (*dto.OrderStatResp, common.Errno) {
	list, count, err := s.order.OrderStatistic(ctx, &do.OrderStat{
		DateType:  req.DateType,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		GoodsID:   req.GoodsID,
	})
	if err != nil {
		logger.Error("OrderStatistic OrderStatistic error", zap.Error(err), zap.Any("req", req))
		return nil, common.DatabaseErr.WithErr(err)
	}
	tempList := make([]*dto.OrderStatDto, 0)
	lo.ForEach(list, func(item *do.OrderStatistic, index int) {
		tempList = append(tempList, &dto.OrderStatDto{
			Date:          item.Date,
			OrderAmount:   item.OrderAmount,
			OrderCount:    item.OrderCount,
			PaymentAmount: item.PaymentAmount,
			RefundAmount:  item.RefundAmount,
			RefundCount:   item.RefundCount,
			RevenueAmount: item.PaymentAmount - item.RefundAmount,
		})
	})
	return &dto.OrderStatResp{
		Pager: req.Pager,
		Total: count,
		List:  tempList,
	}, common.OK
}
