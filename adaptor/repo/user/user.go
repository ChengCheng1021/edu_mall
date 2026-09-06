package user

import (
	"context"
	"errors"
	"mall/adaptor"
	"mall/adaptor/repo/model"
	"mall/adaptor/repo/query"
	"mall/consts"
	"mall/service/do"
	"mall/utils/tools"
	"time"

	"github.com/go-redis/redis"
	"github.com/gogf/gf/util/gconv"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type IUser interface {
	MobileCreateUser(ctx context.Context, req *do.MobileCreateUser) (*model.MobileUser, error)
	WechatCreateUser(ctx context.Context, req *do.WechatCreateUser) (int64, error)
	CreateUserByWechatApp(ctx context.Context, req *do.CreateUserByWechatApp) (*model.User, error)
	UpdateUserPassword(ctx context.Context, req *do.UpdateUserPassword) error
	UpdateUserProfile(ctx context.Context, req *do.UpdateUserProfile) error
	BindWechatApp(ctx context.Context, req *do.BindWechatApp) error
	UnbindWechatApp(ctx context.Context, userID int64, appCode int32) error

	GetMobileUsersByUserIds(ctx context.Context, userIds []int64) ([]*model.MobileUser, error)
	GetMobileUserByUserID(ctx context.Context, userId int64) (*model.MobileUser, error)
	GetUserByMobile(ctx context.Context, mobileSha256 string) (*model.MobileUser, error)
	GetUserByID(ctx context.Context, userId int64) (*model.User, error)
	GetWechatUserByUserID(ctx context.Context, userId int64) (*model.WechatUser, error)
	GetWechatUserByUnionID(ctx context.Context, unionID string) (*model.WechatUser, error)
	GetAppUsersByUserID(ctx context.Context, userId int64) ([]*model.AppUser, error)
	GetUserNameMap(ctx context.Context, ids []int64) (map[int64]string, error)

	GetWechatAppUser(ctx context.Context, appCode int32, openID string) (*model.AppUser, error)

	ListUser(ctx context.Context, req *do.ListUser) ([]*model.User, int64, error)
}

type User struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewUser(adaptor adaptor.IAdaptor) *User {
	return &User{
		db:    adaptor.GetDB(),
		redis: adaptor.GetRedis(),
	}
}
func (a *User) MobileCreateUser(ctx context.Context, req *do.MobileCreateUser) (*model.MobileUser, error) {
	timeNow := time.Now()
	mobileUser := &model.MobileUser{
		MobileAes:    req.MobileAes,
		MobileSha256: req.MobileSha256,
		CreateAt:     timeNow,
		UpdateAt:     timeNow,
	}
	err := a.db.Transaction(func(tx *gorm.DB) error {
		// 创建用户基本信息
		user := &model.User{
			NickName:    req.NickName,
			Sex:         req.Sex,
			Password:    "",
			Status:      consts.IsEnable,
			IconKey:     "",
			CreateAt:    timeNow,
			LastLoginAt: timeNow,
			UpdateAt:    timeNow,
		}

		if err := tx.WithContext(ctx).Create(user).Error; err != nil {
			return err
		}
		// 创建手机号用户关联记录
		mobileUser.UserID = user.ID
		if err := tx.WithContext(ctx).Create(mobileUser).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return mobileUser, nil
}
func (a *User) WechatCreateUser(ctx context.Context, req *do.WechatCreateUser) (int64, error) {
	timeNow := time.Now()
	var userID int64
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user := &model.User{
			NickName:    req.NickName,
			Sex:         consts.SexUnknown,
			Password:    "",
			Status:      consts.IsEnable,
			IconKey:     "",
			CreateAt:    timeNow,
			LastLoginAt: timeNow,
			UpdateAt:    timeNow,
		}
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		userID = user.ID
		if req.UnionID != "" {
			wechatUser := &model.WechatUser{
				UserID:   user.ID,
				UnionID:  req.UnionID,
				NickName: req.NickName,
				IconURL:  req.IconUrl,
				CreateAt: timeNow,
				UpdateAt: timeNow,
			}
			if err := tx.Create(wechatUser).Error; err != nil {
				return err
			}
		}
		if req.OpenID != "" {
			appUser := &model.AppUser{
				UserID:   user.ID,
				AppCode:  req.AppCode,
				OpenID:   req.OpenID,
				Status:   consts.IsEnable,
				CreateAt: timeNow,
				UpdateAt: timeNow,
			}
			if err := tx.Create(appUser).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return userID, err
}

func (a *User) CreateUserByWechatApp(ctx context.Context, req *do.CreateUserByWechatApp) (*model.User, error) {
	timeNow := time.Now()
	user := &model.User{
		NickName:    gconv.String(req.AppCode) + req.OpenID[:5],
		Sex:         consts.SexUnknown,
		Password:    "",
		Status:      consts.IsEnable,
		IconKey:     "",
		CreateAt:    timeNow,
		LastLoginAt: timeNow,
		UpdateAt:    timeNow,
	}
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&model.User{}).Create(user).Error
		if err != nil {
			return err
		}
		appUser := &model.AppUser{
			UserID:   user.ID,
			AppCode:  req.AppCode,
			OpenID:   req.OpenID,
			Status:   consts.IsEnable,
			CreateAt: timeNow,
			UpdateAt: timeNow,
		}
		return tx.Model(&model.AppUser{}).Create(appUser).Error
	})
	return user, err
}

func (a *User) UpdateUserPassword(ctx context.Context, req *do.UpdateUserPassword) error {
	qs := query.Use(a.db).User
	_, err := qs.WithContext(ctx).Where(qs.ID.Eq(req.UserID)).Updates(model.User{
		Password: req.NewPassword,
		UpdateAt: time.Now(),
	})
	return err
}

func (a *User) UpdateUserProfile(ctx context.Context, req *do.UpdateUserProfile) error {
	qs := query.Use(a.db).User
	_, err := qs.WithContext(ctx).Where(qs.ID.Eq(req.UserID)).Updates(model.User{
		NickName: req.NickName,
		Sex:      req.Sex,
		IconKey:  req.IconKey,
		UpdateAt: time.Now(),
	})
	return err
}

func (a *User) BindWechatApp(ctx context.Context, req *do.BindWechatApp) error {
	timeNow := time.Now()
	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		wechatQS := query.Use(tx).WechatUser
		appQS := query.Use(tx).AppUser
		if req.UnionID != "" {
			wechatUser, err := wechatQS.WithContext(ctx).Where(wechatQS.UserID.Eq(req.UserID)).First()
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if wechatUser == nil {
				if err := wechatQS.WithContext(ctx).Create(&model.WechatUser{
					UserID:   req.UserID,
					UnionID:  req.UnionID,
					NickName: req.NickName,
					IconURL:  req.IconURL,
					CreateAt: timeNow,
					UpdateAt: timeNow,
				}); err != nil {
					return err
				}
			} else {
				_, err = wechatQS.WithContext(ctx).Where(wechatQS.UserID.Eq(req.UserID)).Updates(model.WechatUser{
					UnionID:  req.UnionID,
					NickName: req.NickName,
					IconURL:  req.IconURL,
					UpdateAt: timeNow,
				})
				if err != nil {
					return err
				}
			}
		}
		appUser, err := appQS.WithContext(ctx).Where(appQS.UserID.Eq(req.UserID), appQS.AppCode.Eq(req.AppCode)).First()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if appUser == nil {
			return appQS.WithContext(ctx).Create(&model.AppUser{
				UserID:   req.UserID,
				AppCode:  req.AppCode,
				OpenID:   req.OpenID,
				Status:   consts.IsEnable,
				CreateAt: timeNow,
				UpdateAt: timeNow,
			})
		}
		_, err = appQS.WithContext(ctx).Where(appQS.UserID.Eq(req.UserID), appQS.AppCode.Eq(req.AppCode)).Updates(model.AppUser{
			OpenID:   req.OpenID,
			Status:   consts.IsEnable,
			UpdateAt: timeNow,
		})
		return err
	})
}

func (a *User) UnbindWechatApp(ctx context.Context, userID int64, appCode int32) error {
	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		appQS := query.Use(tx).AppUser
		wechatQS := query.Use(tx).WechatUser
		if _, err := appQS.WithContext(ctx).Where(appQS.UserID.Eq(userID), appQS.AppCode.Eq(appCode)).Delete(); err != nil {
			return err
		}
		if _, err := wechatQS.WithContext(ctx).Where(wechatQS.UserID.Eq(userID)).Delete(); err != nil {
			return err
		}
		return nil
	})
}

func (a *User) GetUserByMobile(ctx context.Context, mobileSha256 string) (*model.MobileUser, error) {
	qs := query.Use(a.db).MobileUser
	return qs.WithContext(ctx).Where(qs.MobileSha256.Eq(mobileSha256)).First()
}

func (a *User) GetMobileUserByUserID(ctx context.Context, userId int64) (*model.MobileUser, error) {
	qs := query.Use(a.db).MobileUser
	return qs.WithContext(ctx).Where(qs.UserID.Eq(userId)).First()
}

func (a *User) GetMobileUsersByUserIds(ctx context.Context, userIds []int64) ([]*model.MobileUser, error) {
	qs := query.Use(a.db).MobileUser
	return qs.WithContext(ctx).Where(qs.UserID.In(userIds...)).Find()
}

func (a *User) GetUserByID(ctx context.Context, userId int64) (*model.User, error) {
	qs := query.Use(a.db).User
	return qs.WithContext(ctx).Where(qs.ID.Eq(userId)).First()
}

func (a *User) GetAppUsersByUserID(ctx context.Context, userId int64) ([]*model.AppUser, error) {
	qs := query.Use(a.db).AppUser
	return qs.WithContext(ctx).Where(qs.UserID.Eq(userId)).Find()
}

func (a *User) GetWechatUserByUserID(ctx context.Context, userId int64) (*model.WechatUser, error) {
	qs := query.Use(a.db).WechatUser
	return qs.WithContext(ctx).Where(qs.UserID.Eq(userId)).First()
}

func (a *User) GetWechatUserByUnionID(ctx context.Context, unionID string) (*model.WechatUser, error) {
	qs := query.Use(a.db).WechatUser
	return qs.WithContext(ctx).Where(qs.UnionID.Eq(unionID)).First()
}

func (a *User) GetUserNameMap(ctx context.Context, ids []int64) (map[int64]string, error) {
	qs := query.Use(a.db).User
	list, err := qs.WithContext(ctx).Where(qs.ID.In(ids...)).Find()
	if err != nil {
		return nil, err
	}
	return lo.SliceToMap(list, func(item *model.User) (int64, string) {
		return item.ID, item.NickName
	}), err
}

func (a *User) ListUser(ctx context.Context, req *do.ListUser) ([]*model.User, int64, error) {
	qs := query.Use(a.db).User
	mqs := query.Use(a.db).MobileUser
	tx := qs.WithContext(ctx)
	if len(req.UserIds) > 0 {
		tx = tx.Where(qs.ID.In(req.UserIds...))
	}
	if req.MobileSha256 != "" {
		tx = tx.Join(mqs, mqs.UserID.EqCol(qs.ID)).Where(mqs.MobileSha256.Eq(req.MobileSha256))
	}
	if req.Status != 0 {
		tx = tx.Where(qs.Status.Eq(req.Status))
	}
	if req.NickNameKw != "" {
		tx = tx.Where(qs.NickName.Like(tools.GetAllLike(req.NickNameKw)))
	}
	return tx.Order(qs.LastLoginAt.Desc(), qs.ID.Desc()).FindByPage(req.GetOffset(), req.Limit)
}

func (a *User) GetWechatAppUser(ctx context.Context, appCode int32, openID string) (*model.AppUser, error) {
	qs := query.Use(a.db).AppUser
	return qs.WithContext(ctx).Where(qs.AppCode.Eq(appCode), qs.OpenID.Eq(openID)).First()
}
