package token

import (
	"context"
	"mall/common"
	"mall/utils/logger"

	"go.uber.org/zap"
)

func (s *Service) GetLarkUserAccessToken(ctx context.Context, appCode int32, code string, redirectUrl string, scope string, force bool) (*AccessToken, common.Errno) {
	token, err := s.getLarkUserAccessToken(ctx, appCode, code, redirectUrl, scope, force)
	if err != nil {
		logger.Error("GetTianyuAccessToken get access token failed", zap.Error(err), zap.Any("app_code", appCode), zap.Any("force", force))
		return nil, common.ServerErr.WithErr(err)
	}
	return token, common.OK
}

func (s *Service) getLarkUserAccessToken(ctx context.Context, appCode int32, code string, redirectUrl string, scope string, force bool) (*AccessToken, error) {
	getTokenFunc := func() (*AccessToken, error) {
		token, err := s.lark.GetLarkUserAccessToken(ctx, appCode, code, redirectUrl, scope)
		if err != nil {
			logger.Error("getLarkUserAccessToken GetLarkUserAccessToken get access token failed", zap.Error(err), zap.Any("app_code", appCode), zap.Any("force", force))
			return nil, common.ServerErr.WithErr(err)
		}
		return &AccessToken{
			Token:     token.AccessToken,
			ExpiresIn: token.ExpiresIn,
		}, nil
	}
	rpcToken, err := getTokenFunc()
	if err != nil {
		logger.Error("getLarkUserAccessToken getTokenFunc failed", zap.Error(err), zap.Any("app_code", appCode), zap.Any("force", force))
		return nil, common.ServerErr.WithErr(err)
	}
	return rpcToken, nil
}

func (s *Service) GetLarkTenantAccessToken(ctx context.Context, appCode int32, force bool) (*AccessToken, common.Errno) {
	token, err := s.getLarkTenantAccessToken(ctx, appCode, force)
	if err != nil {
		logger.Error("GetLarkTenantAccessToken get access token failed",
			zap.Error(err),
			zap.Any("app_code", appCode),
			zap.Any("force", force))
		return nil, common.ServerErr.WithErr(err)
	}
	return token, common.OK
}

func (s *Service) getLarkTenantAccessToken(ctx context.Context, appCode int32, force bool) (*AccessToken, error) {
	getTokenFunc := func() (*AccessToken, error) {
		token, err := s.lark.GetLarkTenantAccessToken(ctx, appCode)
		if err != nil {
			logger.Error("getLarkTenantAccessToken  get access token failed",
				zap.Error(err),
				zap.Any("app_code", appCode),
				zap.Any("force", force))
			return nil, common.ServerErr.WithErr(err)
		}
		return &AccessToken{
			Token:     token.TenantAccessToken,
			ExpiresIn: token.Expire,
		}, nil
	}

	lockKey := s.lockTokenKeyFmt(appCode)
	cacheKey := s.cacheTokenKeyFmt(appCode)

	token, err := s.getCache(ctx, cacheKey)
	if err != nil {
		logger.Error("getLarkTenantAccessToken getCache failed",
			zap.Error(err),
			zap.Any("app_code", appCode),
			zap.Any("force", force))
	}
	// 缓存没有获取到，或者调用方强刷时，进行调用获取
	if token == nil || force {
		return s.updateToken(ctx, getTokenFunc, lockKey, cacheKey)
	}
	// 缓存获取成功
	return token, nil
}
