package goods

import (
	"mall/adaptor"
	"mall/adaptor/repo/admin"
	"mall/adaptor/repo/goods"
	"mall/adaptor/repo/user"
	"mall/config"
)

type Service struct {
	conf       *config.Config
	lesson     goods.ILesson
	adminUser  admin.IAdminUser
	course     goods.ICourse
	userCourse user.IUserCourse
}

func NewService(adaptor adaptor.IAdaptor) *Service {
	return &Service{
		conf:       adaptor.GetConfig(),
		lesson:     goods.NewLesson(adaptor),
		adminUser:  admin.NewAdminUser(adaptor),
		course:     goods.NewCourse(adaptor),
		userCourse: user.NewUserCourse(adaptor),
	}
}
