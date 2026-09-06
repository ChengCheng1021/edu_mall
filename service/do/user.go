package do

import (
	"mall/common"
)

type MobileCreateUser struct {
	MobileAes    string
	MobileSha256 string
	NickName     string
	Sex          int32
}

type WechatCreateUser struct {
	UnionID  string
	OpenID   string
	AppCode  int32
	NickName string
	IconUrl  string
}

type CreateUserByWechatApp struct {
	AppCode int32
	OpenID  string
}

type ListUser struct {
	common.Pager
	UserIds      []int64
	NickNameKw   string
	MobileSha256 string
	Status       int32
}

type UpdateUserPassword struct {
	UserID      int64
	NewPassword string
}

type UpdateUserProfile struct {
	UserID   int64
	NickName string
	Sex      int32
	IconKey  string
}

type BindWechatApp struct {
	UserID   int64
	AppCode  int32
	OpenID   string
	UnionID  string
	NickName string
	IconURL  string
}

type ListUserCourse struct {
	common.Pager
	UserId  int64
	GoodsID int64
	Valid   bool
}

type BuyCourseGoods struct {
	OrderItemID       int64
	GoodsID           int64
	GoodType          int32
	LearnExpireTime   int64
	ServiceExpireTime int64
}

type CreateUserCourse struct {
	UserId     int64
	OrderID    int64
	BuyTime    int64
	CourseList []BuyCourseGoods
}

type DeleteUserCourse struct {
	UserID       int64
	OrderID      int64
	OrderItemIds []int64
}
