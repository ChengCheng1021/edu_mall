package role

import (
	"context"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"mall/adaptor/repo/model"
	"mall/common"
	"mall/service/do"
	"mall/service/dto"
	"mall/utils/logger"
)

func (s *Service) CreateRole(ctx context.Context, user *common.AdminUser, req *dto.AddRoleReq) (int64, common.Errno) {
	roleID, err := s.adminRole.CreateRole(ctx, &do.AddRole{
		AdminUserID: user.UserID,
		Name:        req.Name,
		Desc:        req.Desc,
	})
	if err != nil {
		logger.Error("CreateRole CreateRole error", zap.Error(err), zap.Any("req", req))
		return 0, common.DatabaseErr.WithErr(err)
	}
	return roleID, common.OK
}

func (s *Service) UpdateRole(ctx context.Context, user *common.AdminUser, req *dto.UpdateRoleReq) common.Errno {
	err := s.adminRole.UpdateRole(ctx, &do.UpdateRole{
		AdminUserID: user.UserID,
		ID:          req.ID,
		Name:        req.Name,
		Desc:        req.Desc,
		Status:      req.Status,
	})
	if err != nil {
		logger.Error("UpdateRole UpdateRole error", zap.Error(err), zap.Any("req", req))
		return common.DatabaseErr.WithErr(err)
	}
	return common.OK
}

func (s *Service) GetMyRoles(ctx context.Context, user *common.AdminUser) ([]*dto.RoleDto, common.Errno) {
	roleList, err := s.adminRole.GetRoleByUserID(ctx, user.UserID)
	if err != nil {
		logger.Error("GetMyRoles GetRoleByUserID error", zap.Error(err), zap.Any("userid", user.UserID))
		return nil, common.DatabaseErr.WithErr(err)
	}
	roleIds := make([]int64, 0)
	lo.ForEach(roleList, func(item *model.AdminUserRole, index int) {
		roleIds = append(roleIds, item.RoleID)
	})
	roleIds = lo.Uniq(roleIds)

	roleMap, err := s.adminRole.GetRoleByIds(ctx, roleIds)
	if err != nil {
		logger.Error("GetMyRoles GetRoleByIds error", zap.Error(err), zap.Any("roleIds", roleIds))
		return nil, common.DatabaseErr.WithErr(err)
	}

	rolePermMap, err := s.adminRole.GetRolePerms(ctx, roleIds)
	if err != nil {
		logger.Error("GetMyRoles GetRolePerms error", zap.Error(err), zap.Any("roleIds", roleIds))
		return nil, common.DatabaseErr.WithErr(err)
	}
	permIds := make([]int64, 0)
	for _, vList := range rolePermMap {
		permIds = append(permIds, vList...)
	}
	permNameMap, err := s.adminPerm.GetPermNameMap(ctx, lo.Uniq(permIds))
	if err != nil {
		logger.Error("GetMyRoles GetPermNameMap error", zap.Error(err), zap.Any("permIds", permIds))
		return nil, common.DatabaseErr.WithErr(err)
	}
	retList := make([]*dto.RoleDto, 0, len(roleList))
	lo.ForEach(roleList, func(item *model.AdminUserRole, index int) {
		role, ok := roleMap[item.RoleID]
		if !ok {
			role = &model.Role{}
		}
		retList = append(retList, &dto.RoleDto{
			ID:       item.ID,
			Name:     role.Name,
			Desc:     role.Desc,
			Status:   role.Status,
			CreateAt: role.CreateAt.UnixMilli(),
			UpdateAt: role.UpdateAt.UnixMilli(),
			Perms: lo.Map(rolePermMap[item.ID], func(item int64, index int) common.IDName {
				return common.IDName{
					ID:   item,
					Name: permNameMap[item],
				}
			}),
		})
	})
	return retList, common.OK
}

func (s *Service) ListRoles(ctx context.Context, req *dto.ListRoleReq) (*dto.ListRoleResp, common.Errno) {
	list, total, err := s.adminRole.ListRoles(ctx, &do.ListRole{
		Pager:  req.Pager,
		NameKw: req.NameKw,
		Status: req.Status,
	})
	if err != nil {
		logger.Error("ListRoles ListRoles error", zap.Error(err), zap.Any("req", req))
		return nil, common.DatabaseErr.WithErr(err)
	}
	roleIds := make([]int64, 0)
	lo.ForEach(list, func(item *model.Role, index int) {
		roleIds = append(roleIds, item.ID)
	})

	rolePermMap, err := s.adminRole.GetRolePerms(ctx, lo.Uniq(roleIds))
	if err != nil {
		logger.Error("ListRoles GetRolePerms error", zap.Error(err), zap.Any("roleIds", roleIds))
		return nil, common.DatabaseErr.WithErr(err)
	}
	permIds := make([]int64, 0)
	for _, vList := range rolePermMap {
		permIds = append(permIds, vList...)
	}
	permNameMap, err := s.adminPerm.GetPermNameMap(ctx, lo.Uniq(permIds))
	if err != nil {
		logger.Error("ListRoles GetPermNameMap error", zap.Error(err), zap.Any("permIds", permIds))
		return nil, common.DatabaseErr.WithErr(err)
	}

	retList := make([]*dto.RoleDto, 0, len(list))
	lo.ForEach(list, func(item *model.Role, index int) {
		perms := make([]common.IDName, 0)
		lo.ForEach(rolePermMap[item.ID], func(item int64, index int) {
			perms = append(perms, common.IDName{
				ID:   item,
				Name: permNameMap[item],
			})
		})
		retList = append(retList, &dto.RoleDto{
			ID:       item.ID,
			Name:     item.Name,
			Desc:     item.Desc,
			Status:   item.Status,
			Perms:    perms,
			CreateAt: item.CreateAt.UnixMilli(),
			UpdateAt: item.UpdateAt.UnixMilli(),
		})
	})
	return &dto.ListRoleResp{
		Pager: req.Pager,
		Total: total,
		List:  retList,
	}, common.OK
}

func (s *Service) SetRolePerms(ctx context.Context, user *common.AdminUser, req *dto.SetRolePermReq) common.Errno {
	err := s.adminRole.SetRolePerms(ctx, req.RoleID, req.PermIDs, user.UserID)
	if err != nil {
		logger.Error("SetRolePerms SetRolePerms error", zap.Error(err), zap.Any("req", req))
		return common.DatabaseErr.WithErr(err)
	}
	return common.OK
}
