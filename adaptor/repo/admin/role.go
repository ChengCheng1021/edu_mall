package admin

import (
	"context"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"mall/adaptor"
	"mall/adaptor/repo/model"
	"mall/adaptor/repo/query"
	"mall/consts"
	"mall/service/do"
	"mall/utils/tools"
	"time"
)

type IAdminRole interface {
	CreateRole(ctx context.Context, req *do.AddRole) (int64, error)
	UpdateRole(ctx context.Context, req *do.UpdateRole) error
	SetRolePerms(ctx context.Context, roleID int64, permIDs []int64, userId int64) error
	GetRolePerms(ctx context.Context, roleIds []int64) (map[int64][]int64, error)
	ListRoles(ctx context.Context, req *do.ListRole) ([]*model.Role, int64, error)
	GetRoleByUserID(ctx context.Context, userID int64) ([]*model.AdminUserRole, error)
	GetRoleByUserIds(ctx context.Context, userIds []int64) (map[int64][]*model.AdminUserRole, error)
	GetRoleByIds(ctx context.Context, roleIds []int64) (map[int64]*model.Role, error)
}

type AdminRole struct {
	db *gorm.DB
}

func NewAdminRole(adaptor adaptor.IAdaptor) *AdminRole {
	return &AdminRole{
		db: adaptor.GetDB(),
	}
}

func (a *AdminRole) CreateRole(ctx context.Context, req *do.AddRole) (int64, error) {
	timeNow := time.Now()
	qs := query.Use(a.db).Role
	addObj := &model.Role{
		Name:     req.Name,
		Desc:     req.Desc,
		CreateAt: timeNow,
		UpdateAt: timeNow,
		UpdateBy: req.AdminUserID,
		CreateBy: req.AdminUserID,
		Status:   consts.IsEnable,
	}
	err := qs.WithContext(ctx).Create(addObj)
	if err != nil {
		return 0, err
	}
	return addObj.ID, nil
}
func (a *AdminRole) UpdateRole(ctx context.Context, req *do.UpdateRole) error {
	qs := query.Use(a.db).Role
	updateMap := map[string]interface{}{
		qs.Name.ColumnName().String():     req.Name,
		qs.Desc.ColumnName().String():     req.Desc,
		qs.UpdateAt.ColumnName().String(): time.Now(),
		qs.UpdateBy.ColumnName().String(): req.AdminUserID,
	}
	_, err := qs.WithContext(ctx).Where(qs.ID.Eq(req.ID)).Updates(updateMap)
	if err != nil {
		return err
	}
	return nil
}

func (a *AdminRole) SetRolePerms(ctx context.Context, roleID int64, permIDs []int64, userId int64) error {
	qs := query.Use(a.db).RolePermission
	timeNow := time.Now()
	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&model.RolePermission{}).Where(qs.RoleID.Eq(roleID)).Delete(&model.RolePermission{}).Error
		if err != nil {
			return err
		}

		rolePerms := make([]*model.RolePermission, 0)
		for _, permId := range permIDs {
			rolePerms = append(rolePerms, &model.RolePermission{
				RoleID:       roleID,
				PermissionID: permId,
				CreateAt:     timeNow,
				UpdateAt:     timeNow,
				CreateBy:     userId,
				UpdateBy:     userId,
			})
		}
		return tx.CreateInBatches(rolePerms, 100).Error
	})
}
func (a *AdminRole) GetRolePerms(ctx context.Context, roleIds []int64) (map[int64][]int64, error) {
	qs := query.Use(a.db).RolePermission
	list, err := qs.WithContext(ctx).Where(qs.RoleID.In(roleIds...)).Find()
	if err != nil {
		return nil, err
	}
	rolePermMaps := make(map[int64][]int64)
	lo.ForEach(list, func(item *model.RolePermission, index int) {
		rolePermMaps[item.RoleID] = append(rolePermMaps[item.RoleID], item.PermissionID)
	})
	return rolePermMaps, nil
}
func (a *AdminRole) ListRoles(ctx context.Context, req *do.ListRole) ([]*model.Role, int64, error) {
	qs := query.Use(a.db).Role
	tx := qs.WithContext(ctx)
	if req.NameKw != "" {
		tx = tx.Where(qs.Name.Like(tools.GetAllLike(req.NameKw)))
	}
	if req.Status != 0 {
		tx = tx.Where(qs.Status.Eq(req.Status))
	}
	return tx.Order(qs.Status.Desc(), qs.CreateAt.Desc()).FindByPage(req.GetOffset(), req.Limit)
}
func (a *AdminRole) GetRoleByUserID(ctx context.Context, userID int64) ([]*model.AdminUserRole, error) {
	qs := query.Use(a.db).AdminUserRole
	return qs.WithContext(ctx).Where(qs.AdminUserID.Eq(userID)).Find()
}

func (a *AdminRole) GetRoleByUserIds(ctx context.Context, userIds []int64) (map[int64][]*model.AdminUserRole, error) {
	qs := query.Use(a.db).AdminUserRole
	list, err := qs.WithContext(ctx).Where(qs.AdminUserID.In(userIds...)).Find()
	if err != nil {
		return nil, err
	}
	return lo.GroupBy(list, func(item *model.AdminUserRole) int64 {
		return item.AdminUserID
	}), nil
}

func (a *AdminRole) GetRoleByIds(ctx context.Context, roleIds []int64) (map[int64]*model.Role, error) {
	qs := query.Use(a.db).Role
	list, err := qs.WithContext(ctx).Where(qs.ID.In(roleIds...)).Find()
	if err != nil {
		return nil, err
	}
	return lo.SliceToMap(list, func(item *model.Role) (int64, *model.Role) {
		return item.ID, item
	}), nil
}
