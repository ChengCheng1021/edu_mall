package perm

import (
	"context"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"mall/common"
	"mall/service/do"
	"mall/service/dto"
	"mall/utils/logger"
)

func (s *Service) CreatePermission(ctx context.Context, user *common.AdminUser, req *dto.AddPermissionReq) (int64, common.Errno) {
	permID, err := s.adminPerm.CreatePermission(ctx, &do.AddPerm{
		AdminUserID: user.UserID,
		Code:        req.Code,
		Name:        req.Name,
		Type:        req.Type,
		Desc:        req.Desc,
		ParentID:    req.ParentID,
		Sort:        req.Sort,
		PagePath:    req.PagePath,
	})
	if err != nil {
		logger.Error("CreatePermission CreatePermission error", zap.Error(err), zap.Any("req", req))
		return 0, common.DatabaseErr.WithErr(err)
	}
	return permID, common.OK
}

func (s *Service) UpdatePermissions(ctx context.Context, user *common.AdminUser, req *dto.UpdatePermissionReq) common.Errno {
	reqList := make([]do.UpdatePerm, 0)
	lo.ForEach(req.List, func(item dto.UpdatePermDto, index int) {
		reqList = append(reqList, do.UpdatePerm{
			ID: item.ID,
			AddPerm: do.AddPerm{
				AdminUserID: user.UserID,
				Code:        item.Code,
				Name:        item.Name,
				Type:        item.Type,
				Desc:        item.Desc,
				ParentID:    lo.Ternary(item.ParentID == 0, -1, item.ParentID),
				Sort:        item.Sort,
				PagePath:    item.PagePath,
			},
		})
	})
	err := s.adminPerm.UpdatePermission(ctx, &do.UpdatePermList{
		List: reqList,
	})
	if err != nil {
		logger.Error("UpdatePermission UpdatePermission error", zap.Error(err), zap.Any("req", req))
		return common.DatabaseErr.WithErr(err)
	}
	return common.OK
}

func (s *Service) DeletePermission(ctx context.Context, user *common.AdminUser, req *dto.DeletePermissionReq) common.Errno {
	err := s.adminPerm.DeletePermission(ctx, &do.DeletePerm{
		ID: req.ID,
	})
	if err != nil {
		logger.Error("DeletePermission DeletePermission error", zap.Error(err), zap.Any("req", req))
		return common.DatabaseErr.WithErr(err)
	}
	return common.OK
}
