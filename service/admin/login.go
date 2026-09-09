package admin

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
	"mall/utils/tools"
	"time"

	"github.com/go-redis/redis"
	"github.com/gogf/gf/util/gconv"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (s *Service) GetSmsVerifyCode(ctx context.Context, req *dto.GetSmsVerifyCodeReq) common.Errno {
	var err error
	//if !s.shouldBypassSlideCaptcha() {
	_, err = s.verify.GetCaptchaTicket(ctx, req.Ticket)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return common.InvalidCaptchaErr
		}
		logger.Error("GetSmsVerifyCode GetCaptchaTicket error", zap.Error(err), zap.String("mobile", req.Mobile))
		return common.RedisErr.WithErr(err)
	}
	//}
	verifyCode := tools.GenValidateCode(4)
	tokenFun := func(ctx context.Context, force bool) (string, error) {
		token, errno := s.token.GetLarkTenantAccessToken(ctx, consts.LarkAppCode, force)
		if errno.NotOk() {
			return "", common.ServerErr.WithErr(errno)
		}
		return token.Token, nil
	}
	err = s.lark.SendLarkMsg(ctx, tokenFun, &do.SendLarkMsg{
		AppCode: consts.LarkAppCode,
		OpenID:  s.conf.BizConf.LarkGroupID,
		IDType:  rpc.LarkChatGroupType,
		Content: fmt.Sprintf("<b>手机验证码通知</b> :手机号：%s 验证码：%s", req.Mobile, verifyCode),
	})
	if err != nil {
		logger.Error("GetSmsVerifyCode SendLarkMsg error", zap.Error(err), zap.String("mobile", req.Mobile))
		return common.ServerErr.WithErr(err)
	}
	err = s.verify.SetVerifyCode(ctx, req.Mobile, req.Scene, verifyCode, 5*time.Minute)
	if err != nil {
		logger.Error("GetSmsVerifyCode SetVerifyCode error", zap.Error(err), zap.String("mobile", req.Mobile))
		return common.RedisErr.WithErr(err)
	}
	return common.OK
}

func (s *Service) checkSmsVerifyCode(ctx context.Context, mobile, scene, verifyCode string) bool {
	getCode, err := s.verify.GetVerifyCode(ctx, mobile, scene)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false
		}
		logger.Error("CheckSmsVerifyCode GetVerifyCode error", zap.Error(err), zap.String("mobile", mobile))
	}
	return getCode == verifyCode
}

func (s *Service) MobileVerifyLogin(ctx context.Context, req *dto.MobileVerifyCodeLoginReq) (*dto.AdminLoginResp, common.Errno) {
	pass := s.checkSmsVerifyCode(ctx, req.Mobile, consts.AdminUserMobileLoginSmsCode, req.VerifyCode)
	if !pass {
		return nil, common.InvalidSmsCodeErr
	}
	adminUser, err := s.adminUser.GetUserByMobile(ctx, req.Mobile)
	if err != nil {
		logger.Error("MobileVerifyLogin GetUserByMobile error", zap.Error(err), zap.String("mobile", req.Mobile))
		return nil, common.DatabaseErr.WithErr(err)
	}
	if adminUser == nil || adminUser.Status != consts.IsEnable {
		return nil, common.AdminUserNotFound
	}
	return s.handleAdminLogin(ctx, adminUser, err)
}

func (s *Service) handleAdminLogin(ctx context.Context, adminUser *model.AdminUser, err error) (*dto.AdminLoginResp, common.Errno) {
	adminUserDto := dto.AdminUserDto{
		UserID:     adminUser.ID,
		Name:       adminUser.Name,
		NickName:   adminUser.NickName,
		Sex:        adminUser.Sex,
		Status:     adminUser.Status,
		Mobile:     adminUser.Mobile,
		LarkOpenID: adminUser.LarkOpenID,
		UpdateAt:   adminUser.UpdateAt.UnixMilli(),
		CreateAt:   adminUser.CreateAt.UnixMilli(),
	}
	tokenUuid := tools.UUIDHex()
	// 处理token
	err = s.processToken(ctx, tokenUuid, &adminUserDto)
	if err != nil {
		logger.Error("MobileVerifyLogin processToken error", zap.Error(err))
		return nil, common.RedisErr.WithErr(err)
	}
	return &dto.AdminLoginResp{
		Token: tokenUuid,
		User:  adminUserDto,
	}, common.OK
}

func (s *Service) processToken(ctx context.Context, token string, adminUser *dto.AdminUserDto) error {
	err := s.verify.SetAdminUserToken(ctx, adminUser.UserID, token, gconv.String(adminUser), consts.AdminUserTokenExpire)
	if err != nil {
		logger.Error("SetAdminUserToken error", zap.Error(err), zap.String("mobile", adminUser.Mobile))
		return err
	}
	return nil
}

func (s *Service) MobilePasswordLogin(ctx context.Context, req *dto.MobilePasswordLoginReq) (*dto.AdminLoginResp, common.Errno) {
	//if !s.shouldBypassSlideCaptcha() {
	_, err := s.verify.GetCaptchaTicket(ctx, req.Ticket)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, common.InvalidCaptchaErr
		}
		logger.Error("MobileLogin GetCaptchaTicket error", zap.Error(err), zap.String("mobile", req.Mobile))
		return nil, common.RedisErr.WithErr(err)
	}
	//}
	adminUser, err := s.adminUser.GetUserByMobile(ctx, req.Mobile)
	if err != nil {
		logger.Error("MobilePasswordLogin GetUserByMobile error", zap.Error(err), zap.String("mobile", req.Mobile))
		return nil, common.DatabaseErr.WithErr(err)
	}
	if adminUser == nil || adminUser.Status != consts.IsEnable {
		return nil, common.InvalidPasswordErr
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
	if adminUser.Password != req.Password {
		return nil, common.InvalidPasswordErr
	}
	_ = s.verify.DeletePasswordErr(ctx, req.Mobile)

	return s.handleAdminLogin(ctx, adminUser, err)
}

func (s *Service) MobilePasswordReset(ctx context.Context, req *dto.MobilePasswordResetReq) common.Errno {
	pass := s.checkSmsVerifyCode(ctx, req.Mobile, consts.AdminUserReSetPasswordSmsCode, req.VerifyCode)
	if !pass {
		return common.InvalidSmsCodeErr
	}
	if req.Password != req.ConfirmPassword {
		return common.ConfirmPasswordErr
	}
	adminUser, err := s.adminUser.GetUserByMobile(ctx, req.Mobile)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.AdminUserNotFound
		}
		logger.Error("MobilePasswordReset GetUserByMobile error", zap.Error(err), zap.String("mobile", req.Mobile))
		return common.DatabaseErr.WithErr(err)
	}
	if adminUser == nil || adminUser.Status != consts.IsEnable {
		return common.AdminUserNotFound
	}
	err = s.adminUser.UpdateUserPassword(ctx, &do.UpdateAdminUserPassword{
		ID:       adminUser.ID,
		Password: req.ConfirmPassword,
	})
	if err != nil {
		logger.Error("MobilePasswordReset UpdateAdminUserPassword error", zap.Error(err), zap.String("mobile", req.Mobile))
		return common.DatabaseErr.WithErr(err)
	}
	if err = s.verify.CleanAdminUserToken(ctx, adminUser.ID); err != nil && !errors.Is(err, redis.Nil) {
		logger.Error("MobilePasswordReset CleanAdminUserToken error", zap.Error(err), zap.Int64("admin_user_id", adminUser.ID))
		return common.RedisErr.WithErr(err)
	}
	return common.OK
}

func (s *Service) LarkQrCodeLogin(ctx context.Context, req *dto.LarkQrCodeLoginReq) (*dto.AdminLoginResp, common.Errno) {
	accessToken, errno := s.token.GetLarkUserAccessToken(ctx, req.AppCode, req.Code, req.RedirectUri, "", false)
	if errno.NotOk() {
		logger.Error("LarkQrCodeLogin GetLarkUserAccessToken error", zap.Error(errno), zap.Any("req", req))
		return nil, common.ServerErr.WithErr(errno)
	}
	larkUserInfo, err := s.lark.GetLarkUserInfo(ctx, accessToken.Token)
	if err != nil {
		logger.Error("LarkQrCodeLogin GetLarkUserInfo error", zap.Error(err), zap.Any("req", req))
		return nil, common.ServerErr.WithErr(err)
	}
	adminUser, err := s.adminUser.GetUserByLarkOpenID(ctx, larkUserInfo.OpenID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Error("LarkQrCodeLogin GetUserByLarkOpenID error", zap.Error(err), zap.Any("req", req))
			return nil, common.InvalidLarkOpenID
		}
		logger.Error("LarkQrCodeLogin GetUserByLarkOpenID error", zap.Error(err), zap.Any("req", req))
		return nil, common.DatabaseErr.WithErr(err)
	}
	if adminUser == nil || adminUser.Status != consts.IsEnable {
		return nil, common.InvalidLarkOpenID
	}
	return s.handleAdminLogin(ctx, adminUser, err)
}
