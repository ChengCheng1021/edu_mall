package user

import (
	"context"
	"errors"
	"mall/adaptor"
	"mall/adaptor/repo/model"
	"mall/adaptor/repo/query"
	"mall/service/do"
	"time"

	"gorm.io/gorm"
)

type IUserCart interface {
	AddGoods(ctx context.Context, req *do.AddGoods) (int64, error)
	RemoveGoods(ctx context.Context, req *do.RemoveGoods) error
	ListGoods(ctx context.Context, req *do.ListGoods) ([]*model.UserCart, int64, error)
}

type UserCart struct {
	db *gorm.DB
}

func NewUserCart(adaptor adaptor.IAdaptor) *UserCart {
	return &UserCart{
		db: adaptor.GetDB(),
	}
}

func (u *UserCart) AddGoods(ctx context.Context, req *do.AddGoods) (int64, error) {
	qs := query.Use(u.db).UserCart
	addObj := &model.UserCart{
		GoodsID:  req.GoodsID,
		UserID:   req.UserID,
		Quantity: 1,
		AddAt:    time.Now(),
	}
	err := u.db.WithContext(ctx).Create(addObj).Error
	if err != nil && !errors.Is(err, gorm.ErrDuplicatedKey) {
		return 0, err
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		first, err := qs.WithContext(ctx).Where(qs.UserID.Eq(req.UserID), qs.GoodsID.Eq(req.GoodsID)).First()
		if err != nil || first == nil {
			return 0, err
		}
		return first.ID, nil
	}
	return addObj.ID, nil

}
func (u *UserCart) RemoveGoods(ctx context.Context, req *do.RemoveGoods) error {
	qs := query.Use(u.db).UserCart
	_, err := qs.WithContext(ctx).Where(qs.UserID.Eq(req.UserID), qs.ID.Eq(req.ID)).Delete()
	return err
}
func (u *UserCart) ListGoods(ctx context.Context, req *do.ListGoods) ([]*model.UserCart, int64, error) {
	qs := query.Use(u.db).UserCart
	return qs.WithContext(ctx).Where(qs.UserID.Eq(req.UserID)).Order(qs.ID.Desc()).FindByPage(req.GetOffset(), req.Limit)

}
