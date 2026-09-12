package order

import (
	"context"
	"encoding/json"
	"mall/adaptor"
	"mall/adaptor/repo/model"
	"mall/adaptor/repo/query"
	"mall/consts"
	"mall/service/do"
	"mall/utils/tools"
	"sort"
	"time"

	"github.com/gogf/gf/util/gconv"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type IOrder interface {
	CreateOrder(ctx context.Context, req *do.CreateOrder) error
	GetOrderByID(ctx context.Context, orderID int64) (*model.Order, error)
	GetOrderItems(ctx context.Context, orderID int64) ([]*model.OrderItem, error)
	GetOrderRefunds(ctx context.Context, orderID int64) ([]*model.OrderRefund, error)
	CancelOrder(ctx context.Context, orderID, cancelBy int64, cancelType int32) error
	UpdateOrderPaySuccess(ctx context.Context, req *do.UpdateOrderPaySuccess) error
	GetOrderList(ctx context.Context, req *do.GetOrderList) ([]*model.Order, int64, error)
	GetOrderItemByOrderIDs(ctx context.Context, orderIds []int64) (map[int64][]*model.OrderItem, error)

	// 订单统计
	OrderStatistic(ctx context.Context, req *do.OrderStat) ([]*do.OrderStatistic, int64, error)
	// 订单退款
	CreatOrderRefund(ctx context.Context, req *do.OrderRefund) (int64, error)
	GetOrderRefundByOrderID(ctx context.Context, orderID int64) (*model.OrderRefund, error)
	UpdateOrderRefundResult(ctx context.Context, req *do.OrderRefundResult) error
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
func (o *Order) GetOrderByID(ctx context.Context, orderID int64) (*model.Order, error) {
	qs := query.Use(o.db).Order
	return qs.WithContext(ctx).Where(qs.ID.Eq(orderID)).First()
}
func (o *Order) GetOrderItems(ctx context.Context, orderID int64) ([]*model.OrderItem, error) {
	qs := query.Use(o.db).OrderItem
	return qs.WithContext(ctx).Where(qs.OrderID.Eq(orderID)).Find()
}
func (o *Order) GetOrderRefunds(ctx context.Context, orderID int64) ([]*model.OrderRefund, error) {
	qs := query.Use(o.db).OrderRefund
	return qs.WithContext(ctx).Where(qs.OrderID.Eq(orderID)).Find()
}
func (o *Order) GetOrderList(ctx context.Context, req *do.GetOrderList) ([]*model.Order, int64, error) {
	qs := query.Use(o.db).Order
	tx := qs.WithContext(ctx)
	if len(req.StatusList) > 0 {
		statusValues := lo.Map(req.StatusList, func(item int32, index int) int32 {
			return item
		})
		tx = tx.Where(qs.Status.In(statusValues...))
	} else if req.Status != 0 {
		tx = tx.Where(qs.Status.Eq(req.Status))
	}
	if req.OrderID != 0 {
		tx = tx.Where(qs.ID.Eq(req.OrderID))
	}
	if req.UserID != 0 {
		tx = tx.Where(qs.UserID.Eq(req.UserID))
	}
	if req.CreateStart != 0 && req.CreateEnd != 0 {
		tx = tx.Where(qs.CreateAt.Between(req.CreateStart, req.CreateEnd))
	}
	if req.PaymentStart != 0 && req.PaymentEnd != 0 {
		tx = tx.Where(qs.PaymentAt.Between(req.PaymentStart, req.PaymentEnd))
	}
	if req.RefundStart != 0 && req.RefundEnd != 0 {
		tx = tx.Where(qs.RefundAt.Between(req.RefundStart, req.RefundEnd))
	}
	if req.GoodsNameKw != "" {
		tx = tx.Where(qs.OrderDesc.Like(tools.GetAllLike(req.GoodsNameKw)))
	}
	return tx.Order(qs.ID.Desc()).FindByPage(req.GetOffset(), req.Limit)
}
func (o *Order) CancelOrder(ctx context.Context, orderID, cancelBy int64, cancelType int32) error {
	qs := query.Use(o.db).Order
	_, err := qs.WithContext(ctx).Where(qs.ID.Eq(orderID)).UpdateSimple(
		qs.Status.Value(consts.OrderStatusCancel),
		qs.CancelBy.Value(cancelBy),
		qs.CancelType.Value(cancelType),
		qs.CancelAt.Value(time.Now().UnixMilli()),
	)
	return err
}
func (o *Order) GetOrderItemByOrderIDs(ctx context.Context, orderIds []int64) (map[int64][]*model.OrderItem, error) {
	qs := query.Use(o.db).OrderItem
	list, err := qs.WithContext(ctx).Where(qs.OrderID.In(orderIds...)).Find()
	if err != nil {
		return nil, err
	}
	retGroup := lo.GroupBy(list, func(item *model.OrderItem) int64 {
		return item.OrderID
	})
	return retGroup, nil
}
func (o *Order) UpdateOrderPaySuccess(ctx context.Context, req *do.UpdateOrderPaySuccess) error {
	qs := query.Use(o.db).Order
	return o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updateMap := map[string]interface{}{
			qs.Status.ColumnName().String():    consts.OrderStatusPayed,
			qs.PaymentAt.ColumnName().String(): req.PaymentAt.UnixMilli(),
			qs.TradeNo.ColumnName().String():   req.TransactionID,
		}
		err := tx.Model(&Order{}).Where(qs.ID.Eq(req.OrderID)).Updates(updateMap).Error
		if err != nil {
			return err
		}
		return req.BenefitFunc()
	})
}

func (o *Order) CreatOrderRefund(ctx context.Context, req *do.OrderRefund) (int64, error) {
	itemIDs := lo.Uniq(req.ItemIds)
	sort.Slice(itemIDs, func(i, j int) bool {
		return itemIDs[i] < itemIDs[j]
	})
	refund := &model.OrderRefund{
		UserID:       req.UserID,
		ApplyUserID:  req.AdminUserID,
		OrderID:      req.OrderID,
		ItemIds:      gconv.String(itemIDs),
		Status:       consts.RefundStatusProcessing,
		Reason:       req.Reason,
		Amount:       req.Amount,
		ApplyAt:      time.Now().UnixMilli(),
		InnerTradeNo: req.OutTradeNo,
	}
	err := o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Create(refund).Error
		if err != nil {
			return err
		}
		return req.RefundFun(ctx)
	})
	return refund.ID, err
}
func (o *Order) GetOrderRefundByOrderID(ctx context.Context, orderID int64) (*model.OrderRefund, error) {
	qs := query.Use(o.db).OrderRefund
	return qs.WithContext(ctx).
		Where(qs.OrderID.Eq(orderID)).
		Order(qs.Status.Asc(), qs.ID.Desc()).
		First()
}
func (o *Order) OrderStatistic(ctx context.Context, req *do.OrderStat) ([]*do.OrderStatistic, int64, error) {
	var list []*do.OrderStatistic
	tx := o.db.WithContext(ctx).
		Model(&model.Order{}).
		Select(
			o.dateSelect(req.DateType) + " AS date," +
				"COUNT(DISTINCT(orders.id)) AS order_count," +
				"SUM(orders.order_amount) AS order_amount," +
				"SUM(orders.payment_amount) AS payment_amount," +
				"SUM(CASE WHEN orders.status = 3 THEN 1 ELSE 0 END) AS refund_count," +
				"SUM(orders.refund_amount) AS refund_amount").
		Where("orders.status != -1")

	if req.GoodsID != 0 {
		existSubQuery := o.db.Model(&model.OrderItem{}).
			Select("1").
			Where("order_items.goods_id = ?", req.GoodsID)
		tx = tx.Where("EXISTS (?)", existSubQuery)
	}
	tx = tx.Where("orders.payment_at BETWEEN ? AND ?", req.StartTime, req.EndTime).
		Group(o.dateSelect(req.DateType)).
		Order(o.dateSelect(req.DateType) + " DESC")

	var count int64
	if err := tx.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	if err := tx.Offset(req.GetOffset()).Limit(req.Limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, count, nil
}

func (o *Order) dateSelect(dateType int32) string {
	switch dateType {
	case consts.DateTypeYear:
		return "DATE_FORMAT(FROM_UNIXTIME(orders.payment_at/1000), '%Y')"
	case consts.DateTypeQuarter:
		return "CONCAT(DATE_FORMAT(FROM_UNIXTIME(orders.payment_at/1000), '%Y'), '-Q', QUARTER(FROM_UNIXTIME(orders.payment_at/1000)))"
	case consts.DateTypeMonth:
		return "DATE_FORMAT(FROM_UNIXTIME(orders.payment_at/1000), '%Y-%m')"
	default:
		return "DATE_FORMAT(FROM_UNIXTIME(orders.payment_at/1000), '%Y-%m-%d')"
	}
}

func (o *Order) UpdateOrderRefundResult(ctx context.Context, req *do.OrderRefundResult) error {
	allRefund := false
	refundQS := query.Use(o.db).OrderRefund
	orderQS := query.Use(o.db).Order
	orderItemQs := query.Use(o.db).OrderItem
	refundList, err := refundQS.WithContext(ctx).Where(refundQS.OrderID.Eq(req.OrderID)).Find()
	if err != nil {
		return err
	}
	orderItems, err := orderItemQs.WithContext(ctx).Where(orderItemQs.OrderID.Eq(req.OrderID)).Find()
	if err != nil {
		return err
	}
	itemIds := make([]int64, 0)
	for _, v := range refundList {
		tempIds := make([]int64, 0)
		json.Unmarshal([]byte(v.ItemIds), &tempIds)
		itemIds = append(itemIds, tempIds...)
	}
	if len(lo.Uniq(itemIds)) == len(orderItems) {
		allRefund = true
	}
	updateMap := map[string]interface{}{
		refundQS.Status.ColumnName().String(): req.Status,
	}
	if req.Status == consts.RefundStatusDone {
		updateMap[refundQS.DoneAt.ColumnName().String()] = req.SuccessTime
		updateMap[refundQS.RefundID.ColumnName().String()] = req.RefundID
	}
	return o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&model.OrderRefund{}).
			Where(refundQS.ID.Eq(req.OrderRefundID)).
			Updates(updateMap).Error
		if err != nil {
			return err
		}
		if req.Status == consts.RefundStatusDone {
			refundList := make([]*model.OrderRefund, 0)
			err = tx.Model(&model.OrderRefund{}).
				Where(refundQS.OrderID.Eq(req.OrderID), refundQS.Status.Eq(consts.RefundStatusDone)).
				Find(&refundList).Error
			if err != nil {
				return err
			}
			refundAmount := lo.SumBy(refundList, func(item *model.OrderRefund) int64 {
				return item.Amount
			})
			orderUpdateMap := map[string]interface{}{
				orderQS.RefundAmount.ColumnName().String(): refundAmount,
				orderQS.RefundAt.ColumnName().String():     req.SuccessTime,
			}
			if refundAmount >= req.OrderPaymentAmount || allRefund {
				orderUpdateMap[orderQS.Status.ColumnName().String()] = consts.OrderStatusRefund
			}
			err = tx.Model(&model.Order{}).
				Where(orderQS.ID.Eq(req.OrderID)).
				Updates(orderUpdateMap).Error
			if err != nil {
				return err
			}
		}
		return req.RefundDeliveryFun(ctx)
	})
}
