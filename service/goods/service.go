package goods

import (
	"mall/adaptor"
	"mall/adaptor/repo/admin"
	"mall/adaptor/repo/goods"
	"mall/adaptor/repo/upload"
	"mall/adaptor/repo/user"
	"mall/adaptor/rpc"
	"mall/config"
)

type Service struct {
	conf       *config.Config
	lesson     goods.ILesson
	adminUser  admin.IAdminUser
	course     goods.ICourse
	userCourse user.IUserCourse
	storage    rpc.IStorage
	upload     upload.IUploadFile
}

func NewService(adaptor adaptor.IAdaptor) *Service {
	return &Service{
		conf:       adaptor.GetConfig(),
		lesson:     goods.NewLesson(adaptor),
		adminUser:  admin.NewAdminUser(adaptor),
		course:     goods.NewCourse(adaptor),
		userCourse: user.NewUserCourse(adaptor),
		storage:    rpc.NewStorage(adaptor),
		upload:     upload.NewUploadFile(adaptor),
	}
}
