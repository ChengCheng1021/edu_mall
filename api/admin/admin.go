package admin

import (
	"mall/adaptor"
	"mall/service/admin"
	"mall/service/goods"
	"mall/service/order"
	"mall/service/perm"
	"mall/service/role"
	"mall/service/storage"
	"mall/service/user"
)

type Ctrl struct {
	adaptor      adaptor.IAdaptor
	user         *admin.Service
	perm         *perm.Service
	role         *role.Service
	course       *goods.Service
	order        *order.Service
	storage      *storage.Service
	customerUser *user.Service
}

func NewCtrl(adaptor adaptor.IAdaptor) *Ctrl {
	return &Ctrl{
		adaptor:      adaptor,
		user:         admin.NewService(adaptor),
		perm:         perm.NewService(adaptor),
		role:         role.NewService(adaptor),
		course:       goods.NewService(adaptor),
		storage:      storage.NewService(adaptor),
		customerUser: user.NewService(adaptor),
		order:        order.NewService(adaptor),
	}
}
