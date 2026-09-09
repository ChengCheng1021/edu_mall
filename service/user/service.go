package user

import (
	"mall/adaptor"
	"mall/adaptor/redis"
	"mall/adaptor/repo/goods"
	"mall/adaptor/repo/user"
	"mall/adaptor/rpc"
	"mall/config"
	"mall/service/dto"
	"mall/service/token"
	"mall/utils/captcha"

	"github.com/wenlng/go-captcha/v2/slide"
)

type Service struct {
	conf       *config.Config
	verify     redis.IVerify
	captcha    slide.Captcha
	lark       rpc.ILark
	user       user.IUser
	token      *token.Service
	userCourse user.IUserCourse
	goods      goods.ICourse
	storage    rpc.IStorage
}

func NewService(adaptor adaptor.IAdaptor) *Service {
	return &Service{
		conf:       adaptor.GetConfig(),
		verify:     redis.NewVerify(adaptor),
		captcha:    captcha.NewSlideCaptcha(),
		lark:       rpc.NewLark(adaptor),
		token:      token.NewService(adaptor),
		user:       user.NewUser(adaptor),
		userCourse: user.NewUserCourse(adaptor),
		goods:      goods.NewCourse(adaptor),
		storage:    rpc.NewStorage(adaptor),
	}
}

func (s *Service) buildSmsVerifyCodeResp(verifyCode string) *dto.SmsVerifyCodeResp {
	//if !s.isDevEnv() || verifyCode == "" {
	//	return nil
	//}
	return &dto.SmsVerifyCodeResp{DebugVerifyCode: verifyCode}
}
