package do

import "mall/common"

type CreateAdminUser struct {
	AdminUserID int64
	Name        string
	NickName    string
	Mobile      string
	Sex         int32
	RoleIds     []int64
}

type UpdateAdminUser struct {
	AdminUserID int64
	ID          int64
	Name        string
	NickName    string
	Sex         int32
	RoleIds     []int64
	Status      int32
}

type UpdateAdminUserPassword struct {
	ID       int64  `json:"id"`
	Password string `json:"password"`
}

type ListAdminUser struct {
	common.Pager
	Name   string
	Mobile string // 手机号过滤
	RoleID int64
	Status int32 // 状态过滤
}
