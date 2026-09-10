package customer

import (
	"mall/adaptor"
	"mall/service/cart"
	"mall/service/goods"
	"mall/service/order"
	"mall/service/user"
)

type Ctrl struct {
	adaptor adaptor.IAdaptor
	user    *user.Service
	course  *goods.Service
	order   *order.Service
	cart    *cart.Service
}

func NewCtrl(adaptor adaptor.IAdaptor) *Ctrl {
	return &Ctrl{
		adaptor: adaptor,
		user:    user.NewService(adaptor),
		order:   order.NewService(adaptor),
		course:  goods.NewService(adaptor),
		cart:    cart.NewService(adaptor),
	}
}
