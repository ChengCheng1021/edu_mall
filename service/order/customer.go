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
	"strings"

	"github.com/go-redis/redis"
	"github.com/gogf/gf/util/gconv"
	"github.com/jinzhu/copier"
	"github.com/samber/lo"
	"github.com/smartwalle/alipay/v3"
	"go.uber.org/zap"
)

func (s *Service) OrderCalcFee(ctx context.Context, user *common.UserInfo, req *dto.OrderCalcFeeReq) (*dto.OrderCalcFeeResp, common.Errno) {
	courseIDs := lo.Uniq(req.CourseIDs)
	if len(courseIDs) == 0 {
		return nil, common.ParamErr.WithMsg("course ids is empty")
	}
	// 根据课程id获取列表
	courseList, err := s.course.GetCourseInfoByIds(ctx, courseIDs)
	if err != nil {
		logger.Error("OrderCalcFee GetCourseInfoByIds error", zap.Error(err), zap.Any("req", req))
		return nil, common.DatabaseErr.WithErr(err)
	}
	if len(courseList) != len(courseIDs) {
		return nil, common.ParamErr.WithMsg("course goods not found")
	}
	var (
		courseFees       = make([]*dto.CourseFeeDto, 0)
		totalDiscountFee int64
		totalFee         int64
		totalPayFee      int64
	)
	// 根据list组装数据
	for _, course := range courseList {
		discountFee := int64(0)
		feeDto := &dto.CourseFeeDto{
			CourseID:    course.ID,
			Price:       course.CoursePrice,
			DiscountFee: discountFee,
			PayFee:      course.CoursePrice,
			GoodsSnap:   course,
		}
		courseFees = append(courseFees, feeDto)
		totalDiscountFee += feeDto.DiscountFee
		totalFee += feeDto.Price
		totalPayFee += feeDto.PayFee
	}
	feeUUID := tools.UUIDHex()
	orderFeeDto := &dto.OrderCalcFeeResp{
		FeeUUID:          feeUUID,
		TotalFee:         totalFee,
		TotalDiscountFee: totalDiscountFee,
		TotalPayFee:      totalPayFee,
		CourseFees:       courseFees,
	}
	if err := s.rdsOrder.SetOrderCalcFee(ctx, feeUUID, gconv.String(orderFeeDto), consts.OrderCalcFeeExpire); err != nil {
		logger.Error("OrderCalcFee SetOrderCalcFee error", zap.Error(err), zap.Any("req", req))
		return nil, common.DatabaseErr.WithErr(err)
	}
	return orderFeeDto, common.OK
}

func (s *Service) OrderPayNow(ctx context.Context, user *common.UserInfo, req *dto.OrderPayNowReq) (*dto.OrderPayNowResp, common.Errno) {
	orderCalcFeeStr, err := s.rdsOrder.GetOrderCalcFee(ctx, req.FeeUUID)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, common.OrderCalcFeeErr
		}
		logger.Error("OrderPayNow GetOrderCalcFee error", zap.Error(err), zap.Any("req", req))
		return nil, common.DatabaseErr.WithErr(err)
	}
	orderFeeDto := &dto.OrderCalcFeeResp{}
	err = json.Unmarshal([]byte(orderCalcFeeStr), orderFeeDto)
	if err != nil {
		logger.Error("OrderPayNow Unmarshal error", zap.Error(err), zap.Any("req", req))
		return nil, common.ServerErr.WithErr(err)
	}
	if orderFeeDto.TotalPayFee == 0 {
		return nil, common.OrderCalcFeeErr
	}
	orderItems := make([]*do.OrderItem, 0)
	lo.ForEach(orderFeeDto.CourseFees, func(courseFee *dto.CourseFeeDto, index int) {
		orderItems = append(orderItems, &do.OrderItem{
			CourseID:    courseFee.CourseID,
			DiscountFee: courseFee.DiscountFee,
			GoodsSnap:   courseFee.GoodsSnap,
			PayFee:      courseFee.PayFee,
		})
	})
	orderID, err := s.idNode.GetNextID()
	if err != nil {
		logger.Error("OrderPayNow GetNextID error", zap.Error(err), zap.Any("req", req))
		return nil, common.ServerErr.WithErr(err)
	}
	resp := &dto.OrderPayNowResp{OrderID: orderID}
	outTradeNo := ""
	aliPayParam, nativeTradeNo := s.assemblyAliPayParam(ctx, orderFeeDto)
	payUrl, err := s.aliPay.PagePrePayOrder(ctx, aliPayParam)

	if err != nil {
		logger.Error("OrderPayNow NativePrePayOrder error", zap.Error(err), zap.Any("req", req))
		return nil, common.ServerErr.WithErr(err)
	}
	defer func() {
		if err != nil && outTradeNo != "" {
			s.rdsOrder.DelOrderPayResult(ctx, orderID)
			s.aliPay.CloseOrder(ctx, alipay.TradeClose{
				OutTradeNo: outTradeNo,
			})
		}
	}()

	outTradeNo = nativeTradeNo
	resp.CodeURL = payUrl.String()

	err = s.order.CreateOrder(ctx, &do.CreateOrder{
		OrderID:     gconv.Int64(orderID),
		UserID:      user.User.ID,
		OrderSource: consts.OrderSourceUser,
		TotalFee:    orderFeeDto.TotalFee,
		TotalPayFee: orderFeeDto.TotalPayFee,
		UserRemark:  req.Remark,
		OutTradeNo:  outTradeNo,
		OrderDesc:   orderFeeDto.GetDescription(),
		Items:       orderItems,
	})
	if err != nil {
		logger.Error("OrderPayNow CreateOrder error", zap.Error(err), zap.Any("req", req))
		return nil, common.ServerErr.WithErr(err)
	}
	if outTradeNo == "" {
		return resp, common.OK
	}

	err = s.rdsOrder.SetOrderPayResult(ctx, orderID)
	if err != nil {
		logger.Error("OrderPayNow SetOrderPayResult error", zap.Error(err), zap.Any("req", req))
		return nil, common.DatabaseErr.WithErr(err)
	}
	err = s.rdsOrder.SetTimeoutOrderCancel(ctx, orderID)
	if err != nil {
		logger.Error("OrderPayNow SetTimeoutOrderCancel error", zap.Error(err), zap.Any("req", req))
		return nil, common.DatabaseErr.WithErr(err)
	}
	return resp, common.OK
}

func (s *Service) assemblyAliPayParam(ctx context.Context, fee *dto.OrderCalcFeeResp) (alipay.TradePagePay, string) {
	payAmount := gconv.String(fee.TotalPayFee)
	outTradeNo := tools.UUIDHex()
	return alipay.TradePagePay{
		Trade: alipay.Trade{
			OutTradeNo: outTradeNo,
			Subject:    fee.GetShortDesc(),
			//Subject:     "支付数据",
			TotalAmount: strings.TrimSuffix(payAmount, "00"),
			ProductCode: "FAST_INSTANT_TRADE_PAY",
		},
	}, outTradeNo
}
func (s *Service) CancelOrder(ctx context.Context, user *common.UserInfo, req *dto.CancelOrderReq) common.Errno {
	return s.cancelOrder(ctx, req.OrderID, user.User.ID, consts.CustomerCancel)
}

func (s *Service) GetUserOrderList(ctx context.Context, user *common.UserInfo, req *dto.GetOrderListReq) (*dto.GetUserOrderListResp, common.Errno) {
	list, count, err := s.order.GetOrderList(ctx, &do.GetOrderList{
		Pager:       req.Pager,
		Status:      req.Status,
		OrderID:     req.OrderID,
		UserID:      user.User.ID,
		CreateStart: req.CreateStart,
		CreateEnd:   req.CreateEnd,
		GoodsNameKw: req.GoodsNameKw,
	})
	if err != nil {
		logger.Error("GetUserOrderList error", zap.Error(err), zap.Any("req", req))
		return nil, common.DatabaseErr.WithErr(err)
	}
	orderIds := lo.Map(list, func(item *model.Order, index int) int64 {
		return item.ID
	})
	orderItemsMap, err := s.order.GetOrderItemByOrderIDs(ctx, orderIds)
	if err != nil {
		logger.Error("GetUserOrderList GetOrderItemByOrderIDs error", zap.Error(err), zap.Any("req", req))
		return nil, common.DatabaseErr.WithErr(err)
	}
	fileKeys := make([]string, 0)
	goodsSnapMap := make(map[int64]*do.GoodsSnap)
	for _, items := range orderItemsMap {
		lo.ForEach(items, func(item *model.OrderItem, index int) {
			goodsSnap := &do.GoodsSnap{}
			_ = json.Unmarshal([]byte(item.GoodsSnap), goodsSnap)
			fileKeys = append(fileKeys, goodsSnap.CoverKey, goodsSnap.DetailCoverKey)
			goodsSnapMap[item.ID] = goodsSnap
		})
	}
	fileUrlMap, err := s.storage.GetPreviewUrl(ctx, &do.GetPreviewUrl{
		Keys:        fileKeys,
		ExpireHours: 6,
	})
	if err != nil {
		logger.Error("GetUserOrderList GetPreviewUrl error", zap.Error(err), zap.Any("req", req))
		return nil, common.ServerErr.WithErr(err)
	}
	for k, v := range goodsSnapMap {
		goodsSnapMap[k].CoverUrl = fileUrlMap[v.CoverKey]
		goodsSnapMap[k].DetailCoverUrl = fileUrlMap[v.DetailCoverKey]
	}
	orderList := s.convertModelOrderToOrderDto(ctx, list)
	orderMap := lo.SliceToMap(orderList, func(item *dto.OrderDto) (int64, *dto.OrderDto) {
		return item.ID, item
	})
	retList := make([]*dto.OrderInfoResp, 0)
	lo.ForEach(list, func(order *model.Order, index int) {
		modelItems, ok := orderItemsMap[order.ID]
		orderItems := make([]*dto.OrderItemDto, 0)
		if ok {
			copier.Copy(&orderItems, modelItems)
		}
		for i, item := range orderItems {
			orderItems[i].GoodsSnap = goodsSnapMap[item.ID]
		}
		retList = append(retList, &dto.OrderInfoResp{
			OrderDto: orderMap[order.ID],
			Items:    orderItems,
		})
	})
	return &dto.GetUserOrderListResp{
		List:  retList,
		Total: count,
		Pager: req.Pager,
	}, common.OK
}
