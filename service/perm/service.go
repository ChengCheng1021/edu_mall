package perm

import (
	"context"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"mall/adaptor"
	"mall/adaptor/repo/admin"
	"mall/adaptor/repo/model"
	"mall/common"
	"mall/service/dto"
	"mall/utils/logger"
)

type Service struct {
	adminPerm admin.IPerm
}

func NewService(adaptor adaptor.IAdaptor) *Service {
	return &Service{
		adminPerm: admin.NewAdminPerm(adaptor),
	}
}

func (s *Service) PermissionList(ctx context.Context) (*dto.PermissionListResp, common.Errno) {
	permList, total, err := s.adminPerm.PermissionList(ctx, common.Pager{
		Page:      1,
		Limit:     1000,
		UnLimited: true,
	})
	if err != nil {
		logger.Error("PermissionList PermissionList error", zap.Error(err))
		return nil, common.DatabaseErr.WithErr(err)
	}
	retList := make([]*dto.PermissionDto, 0, len(permList))
	lo.ForEach(permList, func(item *model.Permission, index int) {
		retList = append(retList, &dto.PermissionDto{
			ID:       item.ID,
			Code:     item.Code,
			Name:     item.Name,
			Desc:     item.Desc,
			PagePath: item.PagePath,
			ParentID: item.ParentID,
			Sort:     item.Sort,
			Status:   item.Status,
			Type:     item.Type,
			UpdateAt: item.UpdateAt.UnixMilli(),
		})
	})
	return &dto.PermissionListResp{
		Pager: common.Pager{
			Page:      1,
			Limit:     1000,
			UnLimited: true,
		},
		Total: total,
		List:  retList,
	}, common.OK
}

func (s *Service) MyPermissionList(ctx context.Context, user *common.AdminUser) ([]*dto.PermissionDto, common.Errno) {
	permList, err := s.adminPerm.MyPermissionList(ctx, user.UserID)
	if err != nil {
		logger.Error("PermissionList PermissionList error", zap.Error(err))
		return nil, common.DatabaseErr.WithErr(err)
	}
	retList := make([]*dto.PermissionDto, 0, len(permList))
	lo.ForEach(permList, func(item *model.Permission, index int) {
		retList = append(retList, &dto.PermissionDto{
			ID:       item.ID,
			Code:     item.Code,
			Name:     item.Name,
			Desc:     item.Desc,
			PagePath: item.PagePath,
			ParentID: item.ParentID,
			Sort:     item.Sort,
			Status:   item.Status,
			Type:     item.Type,
			UpdateAt: item.UpdateAt.UnixMilli(),
		})
	})
	return retList, common.OK
}
