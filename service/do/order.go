package do

import (
	"context"
	"mall/adaptor/repo/model"
	"mall/common"
	"time"
)

type GoodsSnap struct {
	*model.CourseGood
	CoverUrl       string `json:"cover_url"`
	DetailCoverUrl string `json:"detail_cover_url"`
}

type CreateOrder struct {
	OrderID          int64
	UserID           int64
	TotalDiscountFee int64
	OrderSource      int32
	TotalFee         int64
	TotalPayFee      int64
	UserRemark       string
	Items            []*OrderItem
	OutTradeNo       string
	OrderDesc        string
}

type OrderItem struct {
	CourseID    int64
	DiscountFee int64 // 优惠金额
	PayFee      int64 // 实际支付金额
	GoodsSnap   interface{}
}

type UpdateOrderPaySuccess struct {
	OrderID       int64
	TradeType     string
	TransactionID string
	PaymentAt     time.Time
	BenefitFunc   func() error
}

type GetOrderList struct {
	common.Pager
	Status       int32
	StatusList   []int32
	OrderID      int64
	UserID       int64
	GoodsNameKw  string
	CreateStart  int64
	CreateEnd    int64
	PaymentStart int64
	PaymentEnd   int64
	RefundStart  int64
	RefundEnd    int64
}

type OrderRefund struct {
	UserID      int64
	OrderID     int64
	ItemIds     []int64
	Reason      string
	Amount      int64
	AdminUserID int64
	OutTradeNo  string
	RefundFun   func(ctx context.Context) error
}

type OrderRefundResult struct {
	OrderRefundID      int64
	OrderID            int64
	Status             int32
	RefundID           string
	OrderPaymentAmount int64
	SuccessTime        int64
	RefundDeliveryFun  func(ctx context.Context) error
}

type OrderStat struct {
	common.Pager
	DateType  int32 // 1:按天 2:按月 3：按季度 4:按年
	GoodsID   int64
	StartTime int64
	EndTime   int64
}

type OrderStatistic struct {
	Date          string `gorm:"column:date"`
	OrderCount    int64  `gorm:"column:order_count"`
	OrderAmount   int64  `gorm:"column:order_amount"`
	PaymentAmount int64  `gorm:"column:payment_amount"`
	RefundCount   int64  `gorm:"column:refund_count"`
	RefundAmount  int64  `gorm:"column:refund_amount"`
}
