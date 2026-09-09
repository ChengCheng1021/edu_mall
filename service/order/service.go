package order

import (
	"mall/adaptor"
	"mall/adaptor/repo/goods"
	"mall/config"
)

type Service struct {
	conf   *config.Config
	course goods.ICourse
}

func NewService(adaptor adaptor.IAdaptor) *Service {
	return &Service{
		conf:   adaptor.GetConfig(),
		course: goods.NewCourse(adaptor),
	}
}
