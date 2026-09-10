package order

import (
	"context"
	"mall/adaptor"
	"mall/adaptor/repo/model"
	"mall/consts"
	"mall/service/do"
	"time"

	"github.com/gogf/gf/util/gconv"
	"gorm.io/gorm"
)

type IOrder interface {
	CreateOrder(ctx context.Context, req *do.CreateOrder) error
}

type Order struct {
	db *gorm.DB
}

func NewOrder(adaptor adaptor.IAdaptor) *Order {
	return &Order{
		db: adaptor.GetDB(),
	}
}

func (o *Order) CreateOrder(ctx context.Context, req *do.CreateOrder) error {
	timeNow := time.Now()
	order := &model.Order{
		ID:             req.OrderID,
		UserID:         req.UserID,
		Status:         consts.OrderStatusWaitPay,
		OrderSource:    req.OrderSource,
		OrderAmount:    req.TotalFee,
		PaymentAmount:  req.TotalPayFee,
		TradeNo:        "",
		InnerTradeNo:   req.OutTradeNo,
		OrderDesc:      req.OrderDesc,
		DiscountAmount: req.TotalDiscountFee,
		UserRemark:     req.UserRemark,
		CreateAt:       timeNow.UnixMilli(),
		CreateBy:       req.UserID,
	}
	orderItems := make([]*model.OrderItem, 0)
	for _, v := range req.Items {
		orderItems = append(orderItems, &model.OrderItem{
			OrderID:        order.ID,
			GoodsID:        v.CourseID,
			GoodsType:      consts.CourseGood,
			DiscountAmount: v.DiscountFee,
			PaymentAmount:  v.PayFee,
			GoodsSnap:      gconv.String(v.GoodsSnap),
			Quantity:       1,
			UserID:         req.UserID,
		})
	}

	return o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Create(order).Error
		if err != nil {
			return err
		}
		return tx.Create(orderItems).Error
	})
}
