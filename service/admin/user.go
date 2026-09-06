package admin

import (
	"context"
	"encoding/json"
	"errors"
	"mall/adaptor/repo/model"
	"mall/common"
	"mall/consts"
	"mall/service/do"
	"mall/service/dto"
	"mall/utils/logger"

	"github.com/samber/lo"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (s *Service) GetAdminUserByToken(ctx context.Context, token string) (*common.AdminUser, common.Errno) {
	userString, err := s.verify.GetAdminUserToken(ctx, token)
	if err != nil {
		logger.Error("GetAdminUserByToken GetAdminUserToken", zap.Error(err), zap.Any("token", token))
		return nil, common.RedisErr.WithErr(err)
	}
	adminUser := &common.AdminUser{}
	err = json.Unmarshal([]byte(userString), adminUser)
	if err != nil {
		logger.Error("GetAdminUserByToken json.Unmarshal", zap.Error(err), zap.String("userString", userString))
		return nil, common.ServerErr.WithErr(err)
	}
	if adminUser.UserID == 0 {
		return nil, common.AuthErr
	}
	dbUser, err := s.adminUser.GetUserInfo(ctx, adminUser.UserID)
	if err != nil {
		logger.Error("GetAdminUserByToken GetUserInfo", zap.Error(err), zap.Int64("user_id", adminUser.UserID))
		return nil, common.DatabaseErr.WithErr(err)
	}
	if dbUser == nil || dbUser.Status != consts.IsEnable || dbUser.UpdateAt.UnixMilli() != adminUser.UpdateAt {
		return nil, common.AuthErr
	}
	return adminUser, common.OK
}

func (s *Service) AdminUserLogout(ctx context.Context, adminUser *common.AdminUser) common.Errno {
	err := s.verify.CleanAdminUserToken(ctx, adminUser.UserID)
	if err != nil {
		logger.Error("AdminUserLogout error", zap.Error(err))
		return common.RedisErr.WithErr(err)
	}
	return common.OK
}

func (s *Service) CreateUser(ctx context.Context, adminUser *common.AdminUser, req *dto.CreateUserReq) (int64, common.Errno) {
	userID, err := s.adminUser.CreateUser(ctx, &do.CreateAdminUser{
		AdminUserID: adminUser.UserID,
		Name:        req.Name,
		NickName:    req.NickName,
		Mobile:      req.Mobile,
		Sex:         req.Sex,
		RoleIds:     req.RoleIDs,
	})
	if err != nil {
		logger.Error("CreateAdminUser error", zap.Error(err), zap.Any("req", req))
		return 0, common.DatabaseErr.WithErr(err)
	}
	return userID, common.OK
}

func (s *Service) UpdateUser(ctx context.Context, adminUser *common.AdminUser, req *dto.UpdateUserReq) common.Errno {
	err := s.adminUser.UpdateUser(ctx, &do.UpdateAdminUser{
		ID:          req.ID,
		Name:        req.Name,
		NickName:    req.NickName,
		Sex:         req.Sex,
		Status:      req.Status,
		AdminUserID: adminUser.UserID,
		RoleIds:     req.RoleIDs,
	})
	if err != nil {
		logger.Error("UpdateAdminUser error", zap.Error(err), zap.Any("req", req))
		return common.DatabaseErr.WithErr(err)
	}
	return common.OK
}

func (s *Service) DeleteUser(ctx context.Context, adminUser *common.AdminUser, req *dto.DeleteUserReq) common.Errno {
	err := s.adminUser.DeleteUser(ctx, req.ID)
	if err != nil {
		logger.Error("DeleteUser error", zap.Error(err), zap.Any("req", req))
		return common.DatabaseErr.WithErr(err)
	}
	return common.OK
}

func (s *Service) GetUserInfo(ctx context.Context, adminUser *common.AdminUser) (*dto.AdminUserWithRoleDto, common.Errno) {
	user, err := s.adminUser.GetUserInfo(ctx, adminUser.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.InvalidPasswordErr
		}
		logger.Error("GetUserByID GetUserByID error", zap.Error(err), zap.Any("user_id", adminUser))
		return nil, common.DatabaseErr.WithErr(err)
	}
	userRoleMap, err := s.adminRole.GetRoleByUserIds(ctx, []int64{user.ID})
	if err != nil {
		logger.Error("GetUserByID GetRoleByUserIds error", zap.Error(err), zap.Any("user", user))
		return nil, common.DatabaseErr.WithErr(err)
	}
	roleIds := make([]int64, 0)
	for _, vList := range userRoleMap {
		for _, v := range vList {
			roleIds = append(roleIds, v.RoleID)
		}
	}
	roleMap, err := s.adminRole.GetRoleByIds(ctx, lo.Uniq(roleIds))
	if err != nil {
		logger.Error("GetUserByID GetRoleByIds error", zap.Error(err), zap.Any("roleIds", roleIds))
		return nil, common.DatabaseErr.WithErr(err)
	}
	roles := make([]*common.IDName, 0)
	for _, roleID := range roleIds {
		roles = append(roles, &common.IDName{
			ID:   roleMap[roleID].ID,
			Name: roleMap[roleID].Name,
		})
	}
	return &dto.AdminUserWithRoleDto{
		AdminUserDto: dto.AdminUserDto{
			UserID:     user.ID,
			Name:       user.Name,
			NickName:   user.NickName,
			Sex:        user.Sex,
			Status:     user.Status,
			Mobile:     user.Mobile,
			LarkOpenID: user.LarkOpenID,
			UpdateAt:   user.UpdateAt.UnixMilli(),
			CreateAt:   user.CreateAt.UnixMilli(),
		},
		Roles: roles,
	}, common.OK
}

func (s *Service) AdminUserList(ctx context.Context, adminUser *common.AdminUser, req *dto.ListAdminUserReq) (*dto.ListAdminUserResp, common.Errno) {
	userList, total, err := s.adminUser.ListAdminUser(ctx, &do.ListAdminUser{
		Name:   req.Name,
		Mobile: req.Mobile,
		RoleID: req.RoleID,
		Status: req.Status,
		Pager:  req.Pager,
	})
	if err != nil {
		logger.Error("AdminUserList error", zap.Error(err), zap.Any("req", req))
		return nil, common.DatabaseErr.WithErr(err)
	}
	userIds := make([]int64, 0)
	lo.ForEach(userList, func(item *model.AdminUser, index int) {
		userIds = append(userIds, item.ID)
	})

	userRoleMap, err := s.adminRole.GetRoleByUserIds(ctx, userIds)
	if err != nil {
		logger.Error("AdminUserList GetRoleByUserIds error", zap.Error(err), zap.Any("userIds", userIds))
		return nil, common.DatabaseErr.WithErr(err)
	}
	roleIds := make([]int64, 0)
	for _, vList := range userRoleMap {
		for _, v := range vList {
			roleIds = append(roleIds, v.RoleID)
		}
	}
	roleMap, err := s.adminRole.GetRoleByIds(ctx, lo.Uniq(roleIds))
	if err != nil {
		logger.Error("AdminUserList GetRoleByIds error", zap.Error(err), zap.Any("roleIds", roleIds))
		return nil, common.DatabaseErr.WithErr(err)
	}

	retList := make([]*dto.AdminUserWithRoleDto, 0, len(userList))
	lo.ForEach(userList, func(user *model.AdminUser, index int) {
		retList = append(retList, &dto.AdminUserWithRoleDto{
			AdminUserDto: dto.AdminUserDto{
				ID:         user.ID,
				UserID:     user.ID,
				Name:       user.Name,
				NickName:   user.NickName,
				Sex:        user.Sex,
				Status:     user.Status,
				Mobile:     user.Mobile,
				LarkOpenID: user.LarkOpenID,
				UpdateAt:   user.UpdateAt.UnixMilli(),
				CreateAt:   user.CreateAt.UnixMilli(),
			},
			Roles: lo.Map(userRoleMap[user.ID], func(item *model.AdminUserRole, index int) *common.IDName {
				return &common.IDName{
					ID:   item.RoleID,
					Name: roleMap[item.RoleID].Name,
				}
			}),
		})
	})
	return &dto.ListAdminUserResp{
		List:  retList,
		Total: total,
	}, common.OK
}

func (s *Service) LarkBind(ctx context.Context, adminUser *common.AdminUser, req *dto.LarkQrCodeBindReq) common.Errno {
	accessToken, errno := s.token.GetLarkUserAccessToken(ctx, req.AppCode, req.Code, req.RedirectUri, "", false)
	if errno.NotOk() {
		logger.Error("LarkBind GetLarkUserAccessToken error", zap.Error(errno), zap.Any("req", req))
		return common.ServerErr.WithErr(errno)
	}
	larkUserInfo, err := s.lark.GetLarkUserInfo(ctx, accessToken.Token)
	if err != nil {
		logger.Error("LarkBind GetLarkUserInfo error", zap.Error(err), zap.Any("req", req))
		return common.ServerErr.WithErr(err)
	}

	err = s.adminUser.UpdateUserLarkOpenID(ctx, adminUser.UserID, larkUserInfo.OpenID)
	if err != nil {
		logger.Error("LarkBind UpdateUserLarkOpenID error", zap.Error(err), zap.Any("req", req))
		return common.DatabaseErr.WithErr(err)
	}
	return common.OK
}

func (s *Service) LarkUnbind(ctx context.Context, adminUser *common.AdminUser) common.Errno {
	err := s.adminUser.UpdateUserLarkOpenID(ctx, adminUser.UserID, "")
	if err != nil {
		logger.Error("LarkUnbind error", zap.Error(err), zap.Any("adminUser", adminUser))
		return common.DatabaseErr.WithErr(err)
	}
	return common.OK
}
