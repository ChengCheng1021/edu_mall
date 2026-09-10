package order

import (
	"mall/adaptor"
	"mall/adaptor/redis"
	"mall/adaptor/repo/goods"
	"mall/adaptor/repo/order"
	"mall/adaptor/rpc"
	"mall/config"
)

type Service struct {
	conf     *config.Config
	course   goods.ICourse
	rdsOrder redis.IOrder
	order    order.IOrder
	aliPay   rpc.IPay
}

func NewService(adaptor adaptor.IAdaptor) *Service {
	return &Service{
		conf:     adaptor.GetConfig(),
		course:   goods.NewCourse(adaptor),
		rdsOrder: redis.NewOrder(adaptor),
		order:    order.NewOrder(adaptor),
		aliPay:   rpc.NewAliPay(adaptor),
	}

}
