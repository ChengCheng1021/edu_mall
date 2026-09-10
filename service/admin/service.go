package admin

import (
	"mall/adaptor"
	"mall/adaptor/redis"
	"mall/adaptor/repo/admin"
	"mall/adaptor/repo/user"
	"mall/adaptor/rpc"
	"mall/config"
	"mall/service/token"
	"mall/utils/captcha"
	"strings"

	"github.com/wenlng/go-captcha/v2/slide"
)

type Service struct {
	conf      *config.Config
	adminUser admin.IAdminUser
	adminRole admin.IAdminRole
	user      user.IUser
	verify    redis.IVerify
	captcha   slide.Captcha
	token     *token.Service
	lark      rpc.ILark
}

func NewService(adaptor adaptor.IAdaptor) *Service {
	return &Service{
		conf:      adaptor.GetConfig(),
		adminUser: admin.NewAdminUser(adaptor),
		verify:    redis.NewVerify(adaptor),
		captcha:   captcha.NewSlideCaptcha(),
		user:      user.NewUser(adaptor),
		token:     token.NewService(adaptor),
		lark:      rpc.NewLark(adaptor),
		adminRole: admin.NewAdminRole(adaptor),
	}
}
func (s *Service) shouldBypassSlideCaptcha() bool {
	// 仅 DEV/测试联调使用，当前仅在 DEV 环境开启，生产环境禁止关闭。
	return s != nil && s.conf != nil && strings.EqualFold(s.conf.Server.Env, "dev")
}
