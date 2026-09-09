package user

import (
	"context"
	"errors"
	"fmt"
	"mall/adaptor/repo/model"
	"mall/adaptor/rpc"
	"mall/common"
	"mall/consts"
	"mall/service/do"
	"mall/service/dto"
	"mall/utils/logger"
	"mall/utils/pool"
	"mall/utils/tools"
	"time"

	"github.com/go-redis/redis"
	"github.com/gogf/gf/util/gconv"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (s *Service) GetSmsVerifyCode(ctx context.Context, req *dto.GetSmsVerifyCodeReq) (*dto.SmsVerifyCodeResp, common.Errno) {
	_, err := s.verify.GetCaptchaTicket(ctx, req.Ticket)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, common.InvalidCaptchaErr
		}
		logger.Error("GetSmsVerifyCode GetCaptchaTicket error", zap.Error(err), zap.String("mobile", req.Mobile))
		return nil, common.RedisErr.WithErr(err)
	}

	limited, err := s.CheckSmsLimited(ctx, req.Mobile, req.ClientIP)
	if err != nil {
		logger.Error("CheckSmsLimited error", zap.Error(err), zap.String("mobile", req.Mobile), zap.String("client_ip", req.ClientIP))
		return nil, common.ServerErr.WithErr(err)
	}
	if limited {
		return nil, common.LimitExceedErr
	}

	// 校验手机号是否注册
	if req.Scene == consts.RegisterUserSmsCode || req.Scene == consts.UserReSetPasswordSmsCode {
		mobileSha256 := tools.Sha256Hash(req.Mobile)
		mobileUser, err := s.user.GetUserByMobile(ctx, mobileSha256)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Error("GetSmsVerifyCode GetUserByMobile error", zap.Error(err), zap.String("mobile", req.Mobile))
			return nil, common.DatabaseErr.WithErr(err)
		}
		switch req.Scene {
		case consts.RegisterUserSmsCode:
			if mobileUser != nil {
				return nil, common.MobileRegisteredErr
			}
			// 注册即登录
			req.Scene = consts.UserMobileLoginSmsCode
		case consts.UserReSetPasswordSmsCode:
			if mobileUser == nil {
				return nil, common.UserNotFoundErr
			}
		}
	}
	return s.sendSmsVerifyCode(ctx, req.Mobile, req.Scene)
}

func (s *Service) devSendSms(ctx context.Context, mobile, scene, verifyCode string) common.Errno {
	tokenFun := func(ctx context.Context, force bool) (string, error) {
		token, errno := s.token.GetLarkTenantAccessToken(ctx, consts.LarkAppCode, force)
		if errno.NotOk() {
			return "", common.ServerErr.WithErr(errno)
		}
		return token.Token, nil
	}
	err := s.lark.SendLarkMsg(ctx, tokenFun, &do.SendLarkMsg{
		AppCode: consts.LarkAppCode,
		OpenID:  s.conf.BizConf.LarkGroupID,
		IDType:  rpc.LarkChatGroupType,
		Content: fmt.Sprintf("<b>手机验证码通知</b>\n\n手机号：%s \n验证码：%s", mobile, verifyCode),
	})
	if err != nil {
		logger.Error("sendSmsVerifyCode SendLarkMsg error", zap.Error(err), zap.String("mobile", mobile), zap.String("scene", scene))
		return common.ServerErr.WithErr(err)
	}
	return common.OK
}

func (s *Service) sendSmsVerifyCode(ctx context.Context, mobile, scene string) (*dto.SmsVerifyCodeResp, common.Errno) {
	var (
		errno      = common.OK
		verifyCode = tools.GenValidateCode(4)
	)
	errno = s.devSendSms(ctx, mobile, scene, verifyCode)
	if errno.NotOk() {
		return nil, errno
	}
	err := s.verify.SetVerifyCode(ctx, mobile, scene, verifyCode, 5*time.Minute)
	if err != nil {
		logger.Error("sendSmsVerifyCode SetVerifyCode error", zap.Error(err), zap.String("mobile", mobile), zap.String("scene", scene))
		return nil, common.RedisErr.WithErr(err)
	}
	return s.buildSmsVerifyCodeResp(verifyCode), common.OK
}

func (s *Service) checkSmsVerifyCode(ctx context.Context, mobile, scene, verifyCode string) bool {
	getCode, err := s.verify.GetVerifyCode(ctx, mobile, scene)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false
		}
		logger.Error("CheckSmsVerifyCode GetVerifyCode error", zap.Error(err), zap.String("mobile", mobile))
	}
	if getCode != verifyCode {
		return false
	}
	s.verify.DelVerifyCode(ctx, mobile, scene)
	return true
}

func (s *Service) MobileVerifyLogin(ctx context.Context, req *dto.MobileVerifyCodeLoginReq) (*dto.LoginResp, common.Errno) {
	pass := s.checkSmsVerifyCode(ctx, req.Mobile, consts.UserMobileLoginSmsCode, req.VerifyCode)
	if !pass {
		return nil, common.InvalidSmsCodeErr
	}
	mobileSha256 := tools.Sha256Hash(req.Mobile)
	mobileUser, err := s.user.GetUserByMobile(ctx, mobileSha256)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("MobileVerifyLogin GetUserByMobile error", zap.Error(err), zap.String("mobile", req.Mobile))
		return nil, common.DatabaseErr.WithErr(err)
	}
	// 如果手机号用户不存在，则进行通过手机号创建一个用户
	if mobileUser == nil {
		mobileAes, err := tools.AESEncrypt(req.Mobile, []byte(s.conf.BizConf.MobileSecret))
		if err != nil {
			logger.Error("MobileVerifyLogin AESEncrypt error", zap.Error(err), zap.String("mobile", req.Mobile))
			return nil, common.ServerErr.WithErr(err)
		}
		mobileUser, err = s.user.MobileCreateUser(ctx, &do.MobileCreateUser{
			MobileAes:    mobileAes,    // 密文-用于回显
			MobileSha256: mobileSha256, // 密文-用于查询
			NickName:     fmt.Sprintf("手机号用户%s", req.Mobile[len(req.Mobile)-4:]),
			Sex:          consts.SexUnknown,
		})
		if err != nil {
			logger.Error("MobileVerifyLogin MobileCreateUser error", zap.Error(err), zap.String("mobile", req.Mobile))
			return nil, common.DatabaseErr.WithErr(err)
		}
	}
	user, err := s.user.GetUserByID(ctx, mobileUser.UserID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("MobileVerifyLogin GetUserByID error", zap.Error(err), zap.String("mobile", req.Mobile))
		return nil, common.DatabaseErr.WithErr(err)
	}
	if user == nil || user.Status != consts.IsEnable {
		return nil, common.UserNotFoundErr
	}
	userInfo, err := s.packageUserInfo(ctx, user)
	if err != nil {
		logger.Error("MobileVerifyLogin packageUserInfo error", zap.Error(err), zap.String("mobile", req.Mobile))
		return nil, common.DatabaseErr.WithErr(err)
	}
	return s.handleUserLogin(ctx, userInfo)
}

func (s *Service) packageUserInfo(ctx context.Context, user *model.User) (*dto.UserInfoDto, error) {
	var (
		mobileUser *model.MobileUser
		//appUsers   []*model.AppUser
		//wechatUser *model.WechatUser
		iconURL string
	)
	tempPool := pool.NewPoolWithSize(4)
	defer tempPool.Release()
	tempPool.RunGo(func() {
		tempUser, err := s.user.GetMobileUserByUserID(ctx, user.ID)
		if err != nil {
			logger.Error("MobileVerifyLogin GetMobileUserByUserID error", zap.Error(err), zap.Int64("user_id", user.ID))
			return
		}
		mobileUser = tempUser
	})
	tempPool.RunGo(func() {
		if user.IconKey == "" {
			return
		}
		tempMap, err := s.storage.GetPreviewUrl(ctx, &do.GetPreviewUrl{Keys: []string{user.IconKey}, ExpireHours: 2})
		if err != nil {
			logger.Error("MobileVerifyLogin GetPreviewUrl error", zap.Error(err), zap.Int64("user_id", user.ID), zap.String("icon_key", user.IconKey))
			return
		}
		iconURL = tempMap[user.IconKey]
	})
	tempPool.Wait()
	retUserInfo := &dto.UserInfoDto{
		User: dto.UserDto{
			ID:          user.ID,
			NickName:    user.NickName,
			CreateAt:    user.CreateAt.UnixMilli(),
			IconUrl:     iconURL,
			IconKey:     user.IconKey,
			Sex:         user.Sex,
			Status:      user.Status,
			LastLoginAt: user.LastLoginAt.UnixMilli(),
			UpdateAt:    user.UpdateAt.UnixMilli(),
			HasPassword: user.Password != "",
		},
	}
	if mobileUser != nil {
		mobile, err := tools.AESDecrypt(mobileUser.MobileAes, []byte(s.conf.BizConf.MobileSecret))
		if err != nil {
			logger.Error("MobileVerifyLogin AESDecrypt error", zap.Error(err), zap.String("mobile", mobileUser.MobileAes))
			return nil, common.ServerErr.WithErr(err)
		}
		retUserInfo.MobileUser = &dto.MobileUser{
			Mobile: string(mobile),
			UserID: mobileUser.UserID,
		}
	}
	return retUserInfo, nil
}

func (s *Service) handleUserLogin(ctx context.Context, userInfo *dto.UserInfoDto) (*dto.LoginResp, common.Errno) {
	tokenUuid := tools.UUIDHex()
	// 处理token
	err := s.processToken(ctx, tokenUuid, userInfo)
	if err != nil {
		logger.Error("MobileVerifyLogin processToken error", zap.Error(err))
		return nil, common.RedisErr.WithErr(err)
	}
	return &dto.LoginResp{
		Token:    tokenUuid,
		UserInfo: userInfo,
	}, common.OK
}

func (s *Service) processToken(ctx context.Context, token string, user *dto.UserInfoDto) error {
	err := s.verify.SetUserToken(ctx, user.User.ID, token, gconv.String(user), consts.CustomerUserTokenExpire)
	if err != nil {
		logger.Error("SetAdminUserToken error", zap.Error(err), zap.Any("user", user))
		return err
	}
	return nil
}

func (s *Service) MobilePasswordLogin(ctx context.Context, req *dto.MobilePasswordLoginReq) (*dto.LoginResp, common.Errno) {
	_, err := s.verify.GetCaptchaTicket(ctx, req.Ticket)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, common.InvalidCaptchaErr
		}
		logger.Error("MobileLogin GetCaptchaTicket error", zap.Error(err), zap.String("mobile", req.Mobile))
		return nil, common.RedisErr.WithErr(err)
	}
	mobileSha256 := tools.Sha256Hash(req.Mobile)
	mobileUser, err := s.user.GetUserByMobile(ctx, mobileSha256)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.UserNotFoundErr
		}
		logger.Error("MobilePasswordLogin GetUserByMobile error", zap.Error(err), zap.String("mobile", req.Mobile))
		return nil, common.DatabaseErr.WithErr(err)
	}
	user, err := s.user.GetUserByID(ctx, mobileUser.UserID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("MobilePasswordLogin GetUserByID error", zap.Error(err), zap.String("mobile", req.Mobile))
		return nil, common.DatabaseErr.WithErr(err)
	}
	if user == nil || user.Status != consts.IsEnable {
		return nil, common.UserNotFoundErr
	}
	// 进行用户密码校验累计
	errCount, err := s.verify.IncrPasswordErr(ctx, req.Mobile, consts.PasswordErrExpire)
	if err != nil {
		logger.Error("MobilePasswordLogin IncrPasswordErr error", zap.Error(err), zap.String("mobile", req.Mobile))
		return nil, common.RedisErr.WithErr(err)
	}
	if errCount > consts.PasswordErrMaxCount {
		return nil, common.PasswordErrLimit
	}
	if user.Password != req.Password {
		return nil, common.InvalidPasswordErr
	}
	_ = s.verify.DeletePasswordErr(ctx, req.Mobile)

	userInfo, err := s.packageUserInfo(ctx, user)
	if err != nil {
		logger.Error("MobilePasswordLogin packageUserInfo error", zap.Error(err), zap.String("mobile", req.Mobile))
		return nil, common.DatabaseErr.WithErr(err)
	}
	return s.handleUserLogin(ctx, userInfo)
}

func (s *Service) MobilePasswordReset(ctx context.Context, req *dto.MobilePasswordResetReq) common.Errno {
	pass := s.checkSmsVerifyCode(ctx, req.Mobile, consts.UserReSetPasswordSmsCode, req.VerifyCode)
	if !pass {
		return common.InvalidSmsCodeErr
	}
	if req.Password != req.ConfirmPassword {
		return common.ConfirmPasswordErr
	}
	mobileSha256 := tools.Sha256Hash(req.Mobile)
	mobileUser, err := s.user.GetUserByMobile(ctx, mobileSha256)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.UserNotFoundErr
		}
		logger.Error("MobilePasswordReset GetUserByMobile error", zap.Error(err), zap.String("mobile", req.Mobile))
		return common.DatabaseErr.WithErr(err)
	}
	user, err := s.user.GetUserByID(ctx, mobileUser.UserID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("MobilePasswordReset GetUserByID error", zap.Error(err), zap.String("mobile", req.Mobile))
		return common.DatabaseErr.WithErr(err)
	}
	if user == nil || user.Status != consts.IsEnable {
		return common.UserNotFoundErr
	}

	err = s.user.UpdateUserPassword(ctx, &do.UpdateUserPassword{
		UserID:      user.ID,
		NewPassword: req.ConfirmPassword,
	})
	if err != nil {
		logger.Error("MobilePasswordReset UpdateUserPassword error", zap.Error(err), zap.String("mobile", req.Mobile))
		return common.DatabaseErr.WithErr(err)
	}
	_ = s.verify.CleanUserToken(ctx, user.ID)
	return common.OK
}
