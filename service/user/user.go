package user

import (
	"context"
	"mall/common"
	"mall/consts"
	"mall/service/dto"
	"mall/utils/logger"

	"github.com/goccy/go-json"
	"go.uber.org/zap"
)

func (s *Service) GetUserByToken(ctx context.Context, token string) (*common.UserInfo, error) {
	valueStr, err := s.verify.GetUserToken(ctx, token)
	if err != nil {
		logger.Error("GetUserByToken GetUserToken", zap.Error(err), zap.String("token", token))
		return nil, common.DatabaseErr.WithErr(err)
	}
	userInfo := &common.UserInfo{}
	if err = json.Unmarshal([]byte(valueStr), userInfo); err != nil {
		logger.Error("GetUserByToken json.Unmarshal", zap.Error(err), zap.String("token", token))
		return nil, common.ServerErr.WithErr(err)
	}
	if userInfo.User.ID == 0 {
		return nil, common.AuthErr
	}
	dbUser, err := s.user.GetUserByID(ctx, userInfo.User.ID)
	if err != nil {
		logger.Error("GetUserByToken GetUserByID", zap.Error(err), zap.Int64("user_id", userInfo.User.ID))
		return nil, common.DatabaseErr.WithErr(err)
	}
	if dbUser == nil || dbUser.Status != consts.IsEnable {
		return nil, common.AuthErr
	}
	return userInfo, nil
}

func (s *Service) GetUserInfo(ctx context.Context, userID int64) (*dto.UserInfoDto, common.Errno) {
	user, err := s.user.GetUserByID(ctx, userID)
	if err != nil {
		logger.Error("GetUserInfo GetUserByID", zap.Error(err), zap.Int64("user_id", userID))
		return nil, common.DatabaseErr.WithErr(err)
	}
	userInfo, err := s.packageUserInfo(ctx, user)
	if err != nil {
		logger.Error("GetUserInfo packageUserInfo", zap.Error(err), zap.Int64("user_id", userID))
		return nil, common.DatabaseErr.WithErr(err)
	}
	return userInfo, common.OK
}
