package order

import (
	"context"
	"encoding/json"
	"errors"
	"mall/common"
	"mall/consts"
	"mall/service/do"
	"mall/service/dto"
	"mall/utils/logger"
	"mall/utils/tools"

	"github.com/go-redis/redis"
	"github.com/gogf/gf/util/gconv"
	"github.com/samber/lo"
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

	orderId := tools.UUIDHex()

	err = s.order.CreateOrder(ctx, &do.CreateOrder{
		OrderID:     gconv.Int64(orderId),
		UserID:      user.User.ID,
		OrderSource: consts.OrderSourceUser,
		TotalFee:    orderFeeDto.TotalFee,
		TotalPayFee: orderFeeDto.TotalPayFee,
		UserRemark:  req.Remark,
		OutTradeNo:  "",
		OrderDesc:   orderFeeDto.GetDescription(),
		Items:       orderItems,
	})
	if err != nil {
		logger.Error("OrderPayNow CreateOrder error", zap.Error(err), zap.Any("req", req))
		return nil, common.ServerErr.WithErr(err)
	}
	return &dto.OrderPayNowResp{
		OrderID: gconv.Int64(orderId),
		CodeURL: "https://www.baidu./com",
	}, common.OK
}
