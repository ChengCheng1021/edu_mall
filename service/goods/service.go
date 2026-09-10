package goods

import (
	"mall/adaptor"
	rds "mall/adaptor/redis"
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
	learn      goods.ILearn
	course     goods.ICourse
	userCourse user.IUserCourse
	storage    rpc.IStorage
	upload     upload.IUploadFile
	rdsLearn   rds.ILessonLearn
}

func NewService(adaptor adaptor.IAdaptor) *Service {
	return &Service{
		conf:       adaptor.GetConfig(),
		lesson:     goods.NewLesson(adaptor),
		learn:      goods.NewLearn(adaptor),
		adminUser:  admin.NewAdminUser(adaptor),
		course:     goods.NewCourse(adaptor),
		userCourse: user.NewUserCourse(adaptor),
		storage:    rpc.NewStorage(adaptor),
		upload:     upload.NewUploadFile(adaptor),
		rdsLearn:   rds.NewLessonLearn(adaptor),
	}
}
