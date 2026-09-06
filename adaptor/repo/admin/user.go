package admin

import (
	"context"
	"mall/adaptor"
	"mall/adaptor/repo/model"
	"mall/adaptor/repo/query"
	"mall/consts"
	"mall/service/do"
	"mall/utils/tools"
	"time"

	"github.com/go-redis/redis"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type IAdminUser interface {
	CreateUser(ctx context.Context, req *do.CreateAdminUser) (int64, error)
	UpdateUser(ctx context.Context, req *do.UpdateAdminUser) error
	UpdateUserPassword(ctx context.Context, req *do.UpdateAdminUserPassword) error
	DeleteUser(ctx context.Context, userId int64) error
	UpdateUserLarkOpenID(ctx context.Context, userID int64, openID string) error

	GetUserByMobile(ctx context.Context, mobile string) (*model.AdminUser, error)
	GetUserInfo(ctx context.Context, userId int64) (*model.AdminUser, error)
	GetUserByLarkOpenID(ctx context.Context, openID string) (*model.AdminUser, error)
	GetUserNameMap(ctx context.Context, ids []int64) (map[int64]string, error)

	ListAdminUser(ctx context.Context, req *do.ListAdminUser) ([]*model.AdminUser, int64, error)
}

type AdminUser struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewAdminUser(adaptor adaptor.IAdaptor) *AdminUser {
	return &AdminUser{
		db:    adaptor.GetDB(),
		redis: adaptor.GetRedis(),
	}
}

func (a *AdminUser) CreateUser(ctx context.Context, req *do.CreateAdminUser) (int64, error) {
	timeNow := time.Now()
	addObj := &model.AdminUser{
		Name:     req.Name,
		NickName: req.NickName,
		Mobile:   req.Mobile,
		Sex:      req.Sex,
		CreateAt: timeNow,
		UpdateAt: timeNow,
		UpdateBy: req.AdminUserID,
		Status:   consts.IsEnable,
		CreateBy: req.AdminUserID,
	}
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Create(addObj).Error
		if err != nil {
			return err
		}
		userRoles := make([]model.AdminUserRole, 0)
		for _, roleID := range req.RoleIds {
			userRoles = append(userRoles, model.AdminUserRole{
				AdminUserID: addObj.ID,
				RoleID:      roleID,
				UpdateAt:    timeNow,
				UpdateBy:    req.AdminUserID,
			})
		}
		return tx.CreateInBatches(userRoles, 100).Error
	})
	if err != nil {
		return 0, err
	}
	return addObj.ID, nil
}
func (a *AdminUser) UpdateUser(ctx context.Context, req *do.UpdateAdminUser) error {
	qs := query.Use(a.db).AdminUser
	rqs := query.Use(a.db).AdminUserRole
	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Where(qs.ID.Eq(req.ID)).Updates(model.AdminUser{
			Name:     req.Name,
			NickName: req.NickName,
			Sex:      req.Sex,
			Status:   req.Status,
			UpdateAt: time.Now(),
			UpdateBy: req.AdminUserID,
		}).Error
		if err != nil {
			return err
		}
		err = tx.Where(rqs.AdminUserID.Eq(req.ID)).Delete(model.AdminUserRole{}).Error
		if err != nil {
			return err
		}
		userRoles := make([]model.AdminUserRole, 0)
		for _, roleID := range req.RoleIds {
			userRoles = append(userRoles, model.AdminUserRole{
				AdminUserID: req.ID,
				RoleID:      roleID,
				UpdateAt:    time.Now(),
				UpdateBy:    req.AdminUserID,
			})
		}
		return tx.CreateInBatches(userRoles, 100).Error
	})
}

func (a *AdminUser) UpdateUserPassword(ctx context.Context, req *do.UpdateAdminUserPassword) error {
	qs := query.Use(a.db).AdminUser
	_, err := qs.WithContext(ctx).Where(qs.ID.Eq(req.ID)).Updates(model.AdminUser{
		Password: req.Password,
		UpdateAt: time.Now(),
	})
	if err != nil {
		return err
	}
	return nil
}

func (a *AdminUser) DeleteUser(ctx context.Context, userId int64) error {
	qs := query.Use(a.db).AdminUser
	_, err := qs.WithContext(ctx).Where(qs.ID.Eq(userId)).UpdateSimple(
		qs.IsDelete.Value(consts.IsEnable),
		qs.Mobile.Value(tools.UUIDHex()),
	)
	return err
}

func (a *AdminUser) UpdateUserLarkOpenID(ctx context.Context, userID int64, openID string) error {
	qs := query.Use(a.db).AdminUser
	_, err := qs.WithContext(ctx).Where(qs.ID.Eq(userID)).Update(qs.LarkOpenID, openID)
	return err
}

func (a *AdminUser) GetUserNameMap(ctx context.Context, ids []int64) (map[int64]string, error) {
	qs := query.Use(a.db).AdminUser
	list, err := qs.WithContext(ctx).Where(qs.ID.In(ids...)).Find()
	if err != nil {
		return nil, err
	}
	return lo.SliceToMap(list, func(item *model.AdminUser) (int64, string) {
		return item.ID, item.Name
	}), nil
}

func (a *AdminUser) GetUserInfo(ctx context.Context, userId int64) (*model.AdminUser, error) {
	qs := query.Use(a.db).AdminUser
	return qs.WithContext(ctx).Where(qs.ID.Eq(userId)).First()
}

func (a *AdminUser) GetUserByMobile(ctx context.Context, mobile string) (*model.AdminUser, error) {
	qs := query.Use(a.db).AdminUser
	return qs.WithContext(ctx).Where(qs.Mobile.Eq(mobile)).First()
}

func (a *AdminUser) GetUserByLarkOpenID(ctx context.Context, openID string) (*model.AdminUser, error) {
	qs := query.Use(a.db).AdminUser
	return qs.WithContext(ctx).Where(qs.LarkOpenID.Eq(openID)).First()
}

func (a *AdminUser) ListAdminUser(ctx context.Context, req *do.ListAdminUser) ([]*model.AdminUser, int64, error) {
	qs := query.Use(a.db).AdminUser
	tx := qs.WithContext(ctx).Where(qs.IsDelete.Neq(consts.IsEnable))
	if req.Name != "" {
		tx = tx.Where(qs.Name.Like(tools.GetAllLike(req.Name)))
	}
	if req.Mobile != "" {
		tx = tx.Where(qs.Mobile.Like(tools.GetAllLike(req.Mobile)))
	}
	if req.Status != 0 {
		tx = tx.Where(qs.Status.Eq(req.Status))
	}
	if req.RoleID != 0 {
		rqs := query.Use(a.db).AdminUserRole
		list, err := rqs.WithContext(ctx).Select(rqs.AdminUserID.Distinct()).Where(rqs.RoleID.Eq(req.RoleID)).Find()
		if err != nil {
			return nil, 0, err
		}
		userIds := make([]int64, len(list))
		lo.ForEach(list, func(item *model.AdminUserRole, index int) {
			userIds = append(userIds, item.AdminUserID)
		})
		tx = tx.Where(qs.ID.In(userIds...))
	}
	count, err := tx.Count()
	if err != nil {
		return nil, 0, err
	}
	retList, err := tx.Offset(req.GetOffset()).Limit(req.Limit).Order(qs.Status.Desc(), qs.CreateAt.Desc()).Find()

	return retList, count, err
}
