package cart

import (
	"mall/adaptor"
	"mall/adaptor/repo/goods"
	"mall/adaptor/repo/user"
	"mall/adaptor/rpc"
	"mall/config"
)

type Service struct {
	conf    *config.Config
	cart    user.IUserCart
	course  goods.ICourse
	storage rpc.IStorage
}

func NewService(adaptor adaptor.IAdaptor) *Service {
	return &Service{
		conf:    adaptor.GetConfig(),
		cart:    user.NewUserCart(adaptor),
		course:  goods.NewCourse(adaptor),
		storage: rpc.NewStorage(adaptor),
	}
}
