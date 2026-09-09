package customer

import (
	"mall/adaptor"
	"mall/service/goods"
	"mall/service/user"
)

type Ctrl struct {
	adaptor adaptor.IAdaptor
	user    *user.Service
	course  *goods.Service
}

func NewCtrl(adaptor adaptor.IAdaptor) *Ctrl {
	return &Ctrl{
		adaptor: adaptor,
		user:    user.NewService(adaptor),
		course:  goods.NewService(adaptor),
	}
}
